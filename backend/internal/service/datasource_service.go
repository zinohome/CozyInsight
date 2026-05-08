package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/engine"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type DatasourceService struct {
	repo          *repository.DatasourceRepository
	connectorPool *engine.ConnectorPool
}

func NewDatasourceService(repo *repository.DatasourceRepository, pool *engine.ConnectorPool) *DatasourceService {
	return &DatasourceService{repo: repo, connectorPool: pool}
}

func (s *DatasourceService) Create(ctx context.Context, req *dto.CreateDatasourceRequest, userID uint64) (*model.Datasource, error) {
	// 校验 config 是合法 JSON
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(req.Config), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config json: %w", err)
	}

	ds := &model.Datasource{
		Name:      req.Name,
		Type:      req.Type,
		Config:    req.Config,
		Status:    1,
		CreatedBy: userID,
	}

	if err := s.repo.Create(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DatasourceService) GetByID(ctx context.Context, id uint64) (*model.Datasource, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DatasourceService) List(ctx context.Context) ([]model.Datasource, error) {
	return s.repo.List(ctx)
}

func (s *DatasourceService) Update(ctx context.Context, id uint64, req *dto.UpdateDatasourceRequest) error {
	ds, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		ds.Name = req.Name
	}
	if req.Type != "" {
		ds.Type = req.Type
	}
	if req.Config != "" {
		ds.Config = req.Config
	}
	if req.Status != nil {
		ds.Status = *req.Status
	}

	if err := s.repo.Update(ctx, ds); err != nil {
		return err
	}
	if s.connectorPool != nil {
		s.connectorPool.Remove(id)
	}
	return nil
}

func (s *DatasourceService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if s.connectorPool != nil {
		s.connectorPool.Remove(id)
	}
	return nil
}

func (s *DatasourceService) TestConnection(ctx context.Context, req *dto.TestConnectionRequest) error {
	var dsn string
	switch req.Type {
	case "mysql":
		host, _ := req.Config["host"].(string)
		port, _ := req.Config["port"].(float64)
		username, _ := req.Config["username"].(string)
		password, _ := req.Config["password"].(string)
		database, _ := req.Config["database"].(string)
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%.0f)/%s?charset=utf8mb4&parseTime=true",
			username, password, host, port, database)
	case "postgresql":
		host, _ := req.Config["host"].(string)
		port, _ := req.Config["port"].(float64)
		username, _ := req.Config["username"].(string)
		password, _ := req.Config["password"].(string)
		database, _ := req.Config["database"].(string)
		dsn = fmt.Sprintf("host=%s port=%.0f user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database)
	case "sqlite":
		database, _ := req.Config["database"].(string)
		dsn = database
		req.Type = "sqlite3"
	case "clickhouse":
		host, _ := req.Config["host"].(string)
		port, _ := req.Config["port"].(float64)
		username, _ := req.Config["username"].(string)
		password, _ := req.Config["password"].(string)
		database, _ := req.Config["database"].(string)
		dsn = fmt.Sprintf("clickhouse://%s:%s@%s:%.0f/%s", username, password, host, port, database)
	default:
		return fmt.Errorf("unsupported datasource type: %s", req.Type)
	}

	db, err := sql.Open(req.Type, dsn)
	if err != nil {
		return fmt.Errorf("open connection failed: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}
