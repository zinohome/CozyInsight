package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/engine"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type ChartService struct {
	repo          *repository.ChartRepository
	datasetRepo   *repository.DatasetRepository
	dsRepo        *repository.DatasourceRepository
	cache         *CacheService
	connectorPool *engine.ConnectorPool
}

func NewChartService(repo *repository.ChartRepository, datasetRepo *repository.DatasetRepository, dsRepo *repository.DatasourceRepository, cache *CacheService, pool *engine.ConnectorPool) *ChartService {
	return &ChartService{repo: repo, datasetRepo: datasetRepo, dsRepo: dsRepo, cache: cache, connectorPool: pool}
}

func (s *ChartService) Create(ctx context.Context, req *dto.CreateChartRequest, userID uint64) (*model.Chart, error) {
	if req.Config == "" {
		req.Config = "{}"
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(req.Config), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config json: %w", err)
	}

	chart := &model.Chart{
		Title:     req.Title,
		Type:      req.Type,
		DatasetID: req.DatasetID,
		Config:    req.Config,
		Status:    1,
		CreatedBy: userID,
	}

	if err := s.repo.Create(ctx, chart); err != nil {
		return nil, err
	}
	return chart, nil
}

func (s *ChartService) GetByID(ctx context.Context, id uint64) (*model.Chart, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ChartService) List(ctx context.Context) ([]model.Chart, error) {
	return s.repo.List(ctx)
}

func (s *ChartService) Update(ctx context.Context, id uint64, req *dto.UpdateChartRequest) error {
	chart, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Title != "" {
		chart.Title = req.Title
	}
	if req.Type != "" {
		chart.Type = req.Type
	}
	if req.DatasetID != nil {
		chart.DatasetID = *req.DatasetID
	}
	if req.Config != "" {
		chart.Config = req.Config
	}
	if req.Status != nil {
		chart.Status = *req.Status
	}

	return s.repo.Update(ctx, chart)
}

func (s *ChartService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *ChartService) GetData(ctx context.Context, chartID uint64, runtimeFilters []dto.ChartFilter, drillDimension *string) (*dto.ChartDataResponse, error) {
	chart, err := s.repo.FindByID(ctx, chartID)
	if err != nil {
		return nil, fmt.Errorf("chart not found: %w", err)
	}

	if s.cache != nil {
		if cached, err := s.cache.GetChartData(ctx, chartID, chart.Config); err == nil {
			return cached, nil
		}
	}

	var config dto.ChartConfig
	if err := json.Unmarshal([]byte(chart.Config), &config); err != nil {
		return nil, fmt.Errorf("invalid chart config: %w", err)
	}

	// Append runtime filters
	config.Filters = append(config.Filters, runtimeFilters...)

	// Override dimension for drill-down
	if drillDimension != nil && *drillDimension != "" {
		config.Dimensions = []dto.ChartDimension{{Field: *drillDimension, Sort: "asc"}}
	}

	dataset, err := s.datasetRepo.FindByID(ctx, chart.DatasetID)
	if err != nil {
		return nil, fmt.Errorf("dataset not found: %w", err)
	}

	datasource, err := s.dsRepo.FindByID(ctx, dataset.DatasourceID)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	// Build SQL using query engine
	queryConfig := engine.ChartQueryConfig{
		Dimensions: make([]engine.Dimension, len(config.Dimensions)),
		Metrics:    make([]engine.Metric, len(config.Metrics)),
		Filters:    make([]engine.Filter, len(config.Filters)),
		Orders:     make([]engine.Order, len(config.Orders)),
		Limit:      config.Limit,
	}
	for i, d := range config.Dimensions {
		queryConfig.Dimensions[i] = engine.Dimension{Field: d.Field, Sort: d.Sort}
	}
	for i, m := range config.Metrics {
		queryConfig.Metrics[i] = engine.Metric{Field: m.Field, Aggregation: m.Aggregation, Alias: m.Alias}
	}
	for i, f := range config.Filters {
		queryConfig.Filters[i] = engine.Filter{Field: f.Field, Operator: f.Operator, Value: f.Value}
	}
	for i, o := range config.Orders {
		queryConfig.Orders[i] = engine.Order{Field: o.Field, Direction: o.Direction}
	}

	sql, args, err := engine.BuildSQL(dataset.TableName, datasource.Type, queryConfig)
	if err != nil {
		return nil, fmt.Errorf("build sql failed: %w", err)
	}

	// Execute via connector
	var conn engine.DatasourceConnector
	if s.connectorPool != nil {
		conn, err = s.connectorPool.Get(datasource.ID, datasource.Type, datasource.Config)
	} else {
		conn, err = engine.NewConnector(datasource.Type)
		if err != nil {
			return nil, fmt.Errorf("create connector failed: %w", err)
		}
		if err := conn.Connect(datasource.Config); err != nil {
			return nil, fmt.Errorf("connect failed: %w", err)
		}
		defer conn.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("get connector failed: %w", err)
	}

	data, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Build response metadata
	var dimNames []string
	for _, d := range config.Dimensions {
		dimNames = append(dimNames, d.Field)
	}
	var metricNames []string
	for _, m := range config.Metrics {
		alias := m.Alias
		if alias == "" {
			alias = fmt.Sprintf("%s_%s", strings.ToLower(m.Aggregation), m.Field)
		}
		metricNames = append(metricNames, alias)
	}

	resp := &dto.ChartDataResponse{
		Dimensions: dimNames,
		Metrics:    metricNames,
		Data:       data,
	}

	if s.cache != nil {
		_ = s.cache.SetChartData(ctx, chartID, chart.Config, resp, 5*time.Minute)
	}

	return resp, nil
}
