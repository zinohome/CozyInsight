package service

import (
	"context"
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

	// 新增方法
	Validate(ctx context.Context, id string) error
	GetTables(ctx context.Context, id string) ([]string, error)
	GetTableFields(ctx context.Context, id string, tableName string) ([]map[string]interface{}, error)
}

type datasourceService struct {
	repo repository.DatasourceRepository
}

func NewDatasourceService(repo repository.DatasourceRepository) DatasourceService {
	return &datasourceService{repo: repo}
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

// Validate 验证数据源连接
func (s *datasourceService) Validate(ctx context.Context, id string) error {
	// 获取数据源配置
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("datasource not found: %w", err)
	}

	// 简单验证：检查配置是否完整
	if ds.Type == "" || ds.Name == "" {
		return fmt.Errorf("incomplete datasource configuration")
	}

	// TODO: 实际应该连接数据库测试
	// 这里简化处理，仅返回成功
	return nil
}

// GetTables 获取数据源的表列表
func (s *datasourceService) GetTables(ctx context.Context, id string) ([]string, error) {
	// 获取数据源配置
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	// TODO: 实际应该通过 Calcite 或数据库驱动获取表列表
	// 这里返回模拟数据
	tables := []string{"users", "orders", "products", "categories"}
	return tables, nil
}

// GetTableFields 获取表的字段列表
func (s *datasourceService) GetTableFields(ctx context.Context, id string, tableName string) ([]map[string]interface{}, error) {
	// 获取数据源配置
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	// TODO: 实际应该通过 Calcite 或数据库驱动获取字段列表
	// 这里返回模拟数据
	fields := []map[string]interface{}{
		{"name": "id", "type": "INTEGER"},
		{"name": "name", "type": "VARCHAR"},
		{"name": "created_at", "type": "TIMESTAMP"},
	}
	return fields, nil
}
