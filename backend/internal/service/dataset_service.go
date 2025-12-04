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
	PreviewData(ctx context.Context, id string) ([]map[string]interface{}, error)
}

type datasetService struct {
	repo    repository.DatasetRepository
	calcite *engine.CalciteClient
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

	// If it's a DB table, we might want to fetch fields automatically
	// For now, just save the table definition
	return s.repo.CreateTable(ctx, table)
}

func (s *datasetService) ListTables(ctx context.Context) ([]*model.DatasetTable, error) {
	return s.repo.ListTables(ctx, "") // 空字符串表示获取所有表
}

func (s *datasetService) GetTable(ctx context.Context, id string) (*model.DatasetTable, error) {
	return s.repo.GetTable(ctx, id)
}

func (s *datasetService) PreviewData(ctx context.Context, id string) ([]map[string]interface{}, error) {
	table, err := s.repo.GetTable(ctx, id)
	if err != nil {
		return nil, err
	}

	var sql string
	if table.Type == "sql" {
		// Parse info to get SQL
		var info map[string]interface{}
		if err := json.Unmarshal([]byte(table.Info), &info); err != nil {
			return nil, fmt.Errorf("invalid table info: %w", err)
		}
		if v, ok := info["sql"].(string); ok {
			sql = v
		}
	} else if table.Type == "db" {
		sql = fmt.Sprintf("SELECT * FROM %s LIMIT 100", table.TableName)
	}

	if sql == "" {
		return nil, fmt.Errorf("cannot generate SQL for table %s", table.Name)
	}

	// Use Calcite to execute query
	// Note: In real DataEase, this goes through Avatica which handles the connection to actual datasource
	// Here we assume Avatica is configured to route correctly based on datasource ID
	// But wait, our CalciteClient connects to Avatica Server.
	// The SQL needs to be valid for Calcite.
	// Usually Calcite needs to know about the schema.
	// For this POC, we assume the Avatica server has the datasources configured or we pass connection info.
	// DataEase passes datasourceId in the context or connection properties.
	// For simplicity in this step, we just execute the SQL directly.

	return s.calcite.ExecuteQuery(ctx, sql)
}
