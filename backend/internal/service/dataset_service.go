package service

import (
	"context"
	"cozy-insight-backend/internal/engine"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DatasetService interface {
	CreateGroup(ctx context.Context, group *model.DatasetGroup) error
	ListGroups(ctx context.Context) ([]*model.DatasetGroup, error)

	CreateTable(ctx context.Context, table *model.DatasetTable) error
	ListTables(ctx context.Context) ([]*model.DatasetTable, error)
	GetTable(ctx context.Context, id string) (*model.DatasetTable, error)

	// 数据预览和字段管理
	PreviewData(ctx context.Context, id string, limit int) (*DataPreviewResult, error)
	GetFields(ctx context.Context, id string) ([]*FieldInfo, error)
	SyncFields(ctx context.Context, id string) error
}

type datasetService struct {
	repo    repository.DatasetRepository
	calcite *engine.CalciteClient
}

// DataPreviewResult 数据预览结果
type DataPreviewResult struct {
	Fields []*FieldInfo             `json:"fields"`
	Data   []map[string]interface{} `json:"data"`
	Total  int                      `json:"total"`
}

// FieldInfo 字段信息
type FieldInfo struct {
	Name        string `json:"name"`
	OriginName  string `json:"originName"`
	Type        string `json:"type"`      // 原始数据库类型
	DeType      int    `json:"deType"`    // DataEase类型: 0:text, 1:time, 2:int, 3:float, 4:bool
	GroupType   string `json:"groupType"` // d:维度, q:指标
	Description string `json:"description"`
	Sample      string `json:"sample"` // 示例值
}

func NewDatasetService(repo repository.DatasetRepository, calcite *engine.CalciteClient) DatasetService {
	return &datasetService{
		repo:    repo,
		calcite: calcite,
	}
}

func (s *datasetService) CreateGroup(ctx context.Context, group *model.DatasetGroup) error {
	if group.ID == "" {
		group.ID = uuid.New().String()
	}
	group.CreateTime = time.Now().UnixMilli()
	return s.repo.CreateGroup(ctx, group)
}

func (s *datasetService) ListGroups(ctx context.Context) ([]*model.DatasetGroup, error) {
	return s.repo.ListGroups(ctx)
}

func (s *datasetService) CreateTable(ctx context.Context, table *model.DatasetTable) error {
	if table.ID == "" {
		table.ID = uuid.New().String()
	}
	table.CreateTime = time.Now().UnixMilli()
	table.UpdateTime = time.Now().UnixMilli()
	return s.repo.CreateTable(ctx, table)
}

func (s *datasetService) ListTables(ctx context.Context) ([]*model.DatasetTable, error) {
	return s.repo.ListTables(ctx, "") // 空字符串表示获取所有表
}

func (s *datasetService) GetTable(ctx context.Context, id string) (*model.DatasetTable, error) {
	return s.repo.GetTable(ctx, id)
}

// PreviewData 预览数据集数据
func (s *datasetService) PreviewData(ctx context.Context, id string, limit int) (*DataPreviewResult, error) {
	table, err := s.repo.GetTable(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("table not found: %w", err)
	}

	// 构建 SQL
	sql, err := s.buildPreviewSQL(table, limit)
	if err != nil {
		return nil, err
	}

	// 执行查询
	rows, err := s.calcite.ExecuteQuery(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// 推断字段类型
	fields := s.inferFieldTypes(rows)

	return &DataPreviewResult{
		Fields: fields,
		Data:   rows,
		Total:  len(rows),
	}, nil
}

// GetFields 获取数据集字段信息
func (s *datasetService) GetFields(ctx context.Context, id string) ([]*FieldInfo, error) {
	table, err := s.repo.GetTable(ctx, id)
	if err != nil {
		return nil, err
	}

	// 查询少量数据用于类型推断
	sql, err := s.buildPreviewSQL(table, 10)
	if err != nil {
		return nil, err
	}

	rows, err := s.calcite.ExecuteQuery(ctx, sql)
	if err != nil {
		return nil, err
	}

	return s.inferFieldTypes(rows), nil
}

// SyncFields 同步字段（保存到数据库）
func (s *datasetService) SyncFields(ctx context.Context, id string) error {
	fields, err := s.GetFields(ctx, id)
	if err != nil {
		return err
	}

	// TODO: 保存到 core_dataset_table_field 表
	// 这里简化处理，实际应该通过 repository 保存
	_ = fields

	return nil
}

// buildPreviewSQL 构建预览SQL
func (s *datasetService) buildPreviewSQL(table *model.DatasetTable, limit int) (string, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000 // 最大1000行
	}

	var sql string
	if table.Type == "sql" {
		// 自定义SQL数据集
		var info map[string]interface{}
		if err := json.Unmarshal([]byte(table.Info), &info); err != nil {
			return "", fmt.Errorf("invalid table info: %w", err)
		}
		if v, ok := info["sql"].(string); ok {
			sql = v
		} else {
			return "", fmt.Errorf("sql not found in table info")
		}
	} else if table.Type == "db" {
		// 数据库表
		sql = fmt.Sprintf("SELECT * FROM %s", table.PhysicalTableName)
	} else {
		return "", fmt.Errorf("unsupported table type: %s", table.Type)
	}

	// 添加 LIMIT
	sql = fmt.Sprintf("%s LIMIT %d", sql, limit)

	return sql, nil
}

// inferFieldTypes 推断字段类型
func (s *datasetService) inferFieldTypes(rows []map[string]interface{}) []*FieldInfo {
	if len(rows) == 0 {
		return nil
	}

	var fields []*FieldInfo
	firstRow := rows[0]

	for key, value := range firstRow {
		field := &FieldInfo{
			Name:       key,
			OriginName: key,
		}

		// 推断类型和分组类型
		field.Type, field.DeType, field.GroupType = s.detectFieldType(value)

		// 获取示例值
		if value != nil {
			field.Sample = fmt.Sprintf("%v", value)
			if len(field.Sample) > 50 {
				field.Sample = field.Sample[:50] + "..."
			}
		}

		fields = append(fields, field)
	}

	return fields
}

// detectFieldType 检测字段类型
func (s *datasetService) detectFieldType(value interface{}) (dbType string, deType int, groupType string) {
	if value == nil {
		return "VARCHAR", 0, "d" // 默认文本维度
	}

	switch v := value.(type) {
	case bool:
		return "BOOLEAN", 4, "d" // 布尔类型，维度
	case int, int8, int16, int32, int64:
		return "INTEGER", 2, "q" // 整数，指标
	case uint, uint8, uint16, uint32, uint64:
		return "INTEGER", 2, "q"
	case float32, float64:
		return "DECIMAL", 3, "q" // 小数，指标
	case string:
		// 尝试解析时间
		if s.isTimeString(v) {
			return "TIMESTAMP", 1, "d" // 时间，维度
		}
		return "VARCHAR", 0, "d" // 文本，维度
	case time.Time:
		return "TIMESTAMP", 1, "d"
	case []byte:
		// 字节数组转字符串
		str := string(v)
		if s.isTimeString(str) {
			return "TIMESTAMP", 1, "d"
		}
		return "VARCHAR", 0, "d"
	default:
		return "VARCHAR", 0, "d"
	}
}

// isTimeString 判断是否为时间字符串
func (s *datasetService) isTimeString(str string) bool {
	timeFormats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05",
	}

	for _, format := range timeFormats {
		if _, err := time.Parse(format, str); err == nil {
			return true
		}
	}
	return false
}
