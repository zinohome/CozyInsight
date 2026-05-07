package service

import (
	"context"
	"encoding/json"
	"fmt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type DashboardService struct {
	repo *repository.DashboardRepository
}

func NewDashboardService(repo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) Create(ctx context.Context, req *dto.CreateDashboardRequest, userID uint64) (*model.Dashboard, error) {
	if req.Config == "" {
		req.Config = "{}"
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(req.Config), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config json: %w", err)
	}

	d := &model.Dashboard{
		Title:     req.Title,
		Config:    req.Config,
		Status:    1,
		CreatedBy: userID,
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DashboardService) GetByID(ctx context.Context, id uint64) (*model.Dashboard, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DashboardService) List(ctx context.Context) ([]model.Dashboard, error) {
	return s.repo.List(ctx)
}

func (s *DashboardService) Update(ctx context.Context, id uint64, req *dto.UpdateDashboardRequest) error {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Title != "" {
		d.Title = req.Title
	}
	if req.Config != "" {
		d.Config = req.Config
	}
	if req.Status != nil {
		d.Status = *req.Status
	}

	return s.repo.Update(ctx, d)
}

func (s *DashboardService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DashboardService) AddChart(ctx context.Context, dashboardID uint64, req *dto.AddChartToDashboardRequest) error {
	if req.Config == "" {
		req.Config = "{}"
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(req.Config), &cfg); err != nil {
		return fmt.Errorf("invalid config json: %w", err)
	}

	dc := &model.DashboardChart{
		DashboardID: dashboardID,
		ChartID:     req.ChartID,
		PositionX:   req.PositionX,
		PositionY:   req.PositionY,
		Width:       req.Width,
		Height:      req.Height,
		Config:      req.Config,
	}

	return s.repo.AddChart(ctx, dc)
}

func (s *DashboardService) GetCharts(ctx context.Context, dashboardID uint64) ([]model.DashboardChart, error) {
	return s.repo.GetCharts(ctx, dashboardID)
}

func (s *DashboardService) RemoveChart(ctx context.Context, dashboardID, chartID uint64) error {
	return s.repo.RemoveChart(ctx, dashboardID, chartID)
}
