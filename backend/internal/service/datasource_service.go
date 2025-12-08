package service

import (
	"context"
	"cozy-insight-backend/internal/engine"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DatasourceService interface {
	Create(ctx context.Context, ds *model.Datasource) error
	Update(ctx context.Context, ds *model.Datasource) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*model.Datasource, error)
	List(ctx context.Context) ([]*model.Datasource, error)

	// 连接测试和元数据查询
	TestConnection(ctx context.Context, id string) (*engine.ConnectionTestResult, error)
	TestConnectionByConfig(ctx context.Context, config string) (*engine.ConnectionTestResult, error)
	GetDatabases(ctx context.Context, id string) ([]string, error)
	GetTables(ctx context.Context, id string, database string) ([]string, error)
	GetTableSchema(ctx context.Context, id string, database, table string) ([]map[string]interface{}, error)
}

type datasourceService struct {
	repo      repository.DatasourceRepository
	connector *engine.DatasourceConnector
}

func NewDatasourceService(repo repository.DatasourceRepository, connector *engine.DatasourceConnector) DatasourceService {
	return &datasourceService{
		repo:      repo,
		connector: connector,
	}
}

func (s *datasourceService) Create(ctx context.Context, ds *model.Datasource) error {
	if ds.ID == "" {
		ds.ID = uuid.New().String()
	}
	ds.CreateTime = time.Now().UnixMilli()
	ds.UpdateTime = time.Now().UnixMilli()
	return s.repo.Create(ctx, ds)
}

func (s *datasourceService) Update(ctx context.Context, ds *model.Datasource) error {
	ds.UpdateTime = time.Now().UnixMilli()
	return s.repo.Update(ctx, ds)
}

func (s *datasourceService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *datasourceService) GetByID(ctx context.Context, id string) (*model.Datasource, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *datasourceService) List(ctx context.Context) ([]*model.Datasource, error) {
	return s.repo.List(ctx)
}

// TestConnection 测试数据源连接（通过 ID）
func (s *datasourceService) TestConnection(ctx context.Context, id string) (*engine.ConnectionTestResult, error) {
	// 获取数据源配置
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	// 使用连接器测试连接
	return s.connector.TestConnection(ctx, ds.Configuration)
}

// TestConnectionByConfig 测试数据源连接（通过配置）
func (s *datasourceService) TestConnectionByConfig(ctx context.Context, config string) (*engine.ConnectionTestResult, error) {
	return s.connector.TestConnection(ctx, config)
}

// GetDatabases 获取数据源的数据库列表
func (s *datasourceService) GetDatabases(ctx context.Context, id string) ([]string, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	return s.connector.GetDatabaseList(ctx, ds.Configuration)
}

// GetTables 获取数据源的表列表
func (s *datasourceService) GetTables(ctx context.Context, id string, database string) ([]string, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	return s.connector.GetTableList(ctx, ds.Configuration, database)
}

// GetTableSchema 获取表结构
func (s *datasourceService) GetTableSchema(ctx context.Context, id string, database, table string) ([]map[string]interface{}, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	return s.connector.GetTableSchema(ctx, ds.Configuration, database, table)
}
