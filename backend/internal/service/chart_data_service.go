package service

import (
	"context"
	"cozy-insight-backend/internal/engine"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"cozy-insight-backend/pkg/logger"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type ChartDataService interface {
	GetChartData(ctx context.Context, chartID string) ([]map[string]interface{}, error)
	GetChartDataWithFilter(ctx context.Context, chartID string, filter *QueryFilter) ([]map[string]interface{}, error)
}

type chartDataService struct {
	chartRepo   repository.ChartRepository
	datasetRepo repository.DatasetRepository
	calcite     *engine.CalciteClient
}

// QueryFilter 查询过滤器
type QueryFilter struct {
	Filters []FilterCondition `json:"filters"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
}

// FilterCondition 过滤条件
type FilterCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // =, !=, >, <, >=, <=, LIKE, IN
	Value    interface{} `json:"value"`
}

// AxisConfig 轴配置
type AxisConfig struct {
	Fields []FieldConfig `json:"fields"`
}

// FieldConfig 字段配置
type FieldConfig struct {
	Name      string `json:"name"`
	Aggregate string `json:"aggregate"` // SUM, AVG, COUNT, MAX, MIN
	Sort      string `json:"sort"`      // ASC, DESC
	DataType  string `json:"dataType"`
}

func NewChartDataService(chartRepo repository.ChartRepository, datasetRepo repository.DatasetRepository, calcite *engine.CalciteClient) ChartDataService {
	return &chartDataService{
		chartRepo:   chartRepo,
		datasetRepo: datasetRepo,
		calcite:     calcite,
	}
}

func (s *chartDataService) GetChartData(ctx context.Context, chartID string) ([]map[string]interface{}, error) {
	return s.GetChartDataWithFilter(ctx, chartID, nil)
}

func (s *chartDataService) GetChartDataWithFilter(ctx context.Context, chartID string, filter *QueryFilter) ([]map[string]interface{}, error) {
	// 获取图表配置
	chart, err := s.chartRepo.Get(ctx, chartID)
	if err != nil {
		logger.Log.Error("failed to get chart", zap.String("chartId", chartID), zap.Error(err))
		return nil, fmt.Errorf("chart not found: %w", err)
	}

	// 获取数据集
	table, err := s.datasetRepo.GetTable(ctx, chart.TableID)
	if err != nil {
		logger.Log.Error("failed to get table", zap.String("tableId", chart.TableID), zap.Error(err))
		return nil, fmt.Errorf("dataset not found: %w", err)
	}

	// 构建 SQL
	sql, err := s.buildChartSQL(chart, table, filter)
	if err != nil {
		logger.Log.Error("failed to build SQL", zap.Error(err))
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	logger.Log.Info("executing chart query", zap.String("sql", sql))

	// 执行查询
	data, err := s.executeQuery(ctx, sql)
	if err != nil {
		logger.Log.Error("failed to execute query", zap.Error(err))
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return data, nil
}

// buildChartSQL 构建图表查询 SQL
func (s *chartDataService) buildChartSQL(chart *model.ChartView, dataset *model.DatasetTable, filter *QueryFilter) (string, error) {
	// 解析 X 轴和 Y 轴配置
	var xAxisConfig, yAxisConfig AxisConfig
	if chart.XAxis != "" {
		if err := json.Unmarshal([]byte(chart.XAxis), &xAxisConfig); err != nil {
			return "", fmt.Errorf("invalid xAxis config: %w", err)
		}
	}
	if chart.YAxis != "" {
		if err := json.Unmarshal([]byte(chart.YAxis), &yAxisConfig); err != nil {
			return "", fmt.Errorf("invalid yAxis config: %w", err)
		}
	}

	// 获取基础 SQL
	var baseSQL string
	if dataset.Type == "sql" {
		var info map[string]interface{}
		if err := json.Unmarshal([]byte(dataset.Info), &info); err != nil {
			return "", fmt.Errorf("invalid dataset info: %w", err)
		}
		if v, ok := info["sql"].(string); ok {
			baseSQL = v
		}
	} else if dataset.Type == "db" {
		baseSQL = fmt.Sprintf("SELECT * FROM %s", dataset.PhysicalTableName)
	}

	// 构建 SELECT 子句
	var selectFields []string
	var groupByFields []string

	// X 轴字段（维度）
	for _, field := range xAxisConfig.Fields {
		selectFields = append(selectFields, field.Name)
		groupByFields = append(groupByFields, field.Name)
	}

	// Y 轴字段（指标，需要聚合）
	for _, field := range yAxisConfig.Fields {
		if field.Aggregate != "" {
			selectFields = append(selectFields, fmt.Sprintf("%s(%s) AS %s",
				field.Aggregate, field.Name, field.Name))
		} else {
			selectFields = append(selectFields, field.Name)
		}
	}

	if len(selectFields) == 0 {
		selectFields = append(selectFields, "*")
	}

	// 构建完整 SQL
	sql := fmt.Sprintf("SELECT %s FROM (%s) AS base",
		strings.Join(selectFields, ", "), baseSQL)

	// 添加 WHERE 条件
	if filter != nil && len(filter.Filters) > 0 {
		whereConditions := s.buildWhereConditions(filter.Filters)
		if len(whereConditions) > 0 {
			sql += " WHERE " + strings.Join(whereConditions, " AND ")
		}
	}

	// 添加 GROUP BY
	if len(groupByFields) > 0 {
		sql += " GROUP BY " + strings.Join(groupByFields, ", ")
	}

	// 添加 ORDER BY
	for _, field := range yAxisConfig.Fields {
		if field.Sort != "" {
			sql += fmt.Sprintf(" ORDER BY %s %s", field.Name, field.Sort)
			break // 只取第一个排序字段
		}
	}

	// 添加 LIMIT
	if filter != nil && filter.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			sql += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	} else {
		sql += " LIMIT 1000" // 默认限制
	}

	return sql, nil
}

// extractFields 从配置中提取字段名
func (s *chartDataService) extractFields(config map[string]interface{}) []string {
	var fields []string

	if config == nil {
		return fields
	}

	// 尝试从 fields 数组中提取
	if fieldsData, ok := config["fields"].([]interface{}); ok {
		for _, f := range fieldsData {
			if fieldMap, ok := f.(map[string]interface{}); ok {
				if name, ok := fieldMap["name"].(string); ok {
					fields = append(fields, name)
				}
			}
		}
	}

	return fields
}

// executeQuery 执行 SQL 查询
func (s *chartDataService) executeQuery(ctx context.Context, sql string) ([]map[string]interface{}, error) {
	if s.calcite == nil {
		return nil, fmt.Errorf("calcite engine not initialized")
	}

	return s.calcite.ExecuteQuery(ctx, sql)
}

// joinFields 连接字段名
func joinFields(fields []string) string {
	if len(fields) == 0 {
		return "*"
	}

	result := ""
	for i, field := range fields {
		if i > 0 {
			result += ", "
		}
		result += field
	}
	return result
}
