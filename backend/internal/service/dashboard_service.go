package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type DashboardService struct {
	repo         *repository.DashboardRepository
	shareLinkRepo *repository.ShareLinkRepository
}

func NewDashboardService(repo *repository.DashboardRepository, shareLinkRepo *repository.ShareLinkRepository) *DashboardService {
	return &DashboardService{repo: repo, shareLinkRepo: shareLinkRepo}
}

func (s *DashboardService) Create(ctx context.Context, req *dto.CreateDashboardRequest, userID uint64) (*model.Dashboard, error) {
	if req.Config == "" {
		req.Config = "{}"
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(req.Config), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config json: %w", err)
	}

	if req.Type == "" {
		req.Type = "dashboard"
	}
	if req.Type != "dashboard" && req.Type != "screen" {
		return nil, fmt.Errorf("invalid dashboard type: %s", req.Type)
	}

	d := &model.Dashboard{
		Title:     req.Title,
		Type:      req.Type,
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

func (s *DashboardService) Update(ctx context.Context, id uint64, req *dto.UpdateDashboardRequest, userID uint64) error {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if d.CreatedBy != userID {
		return ErrNotOwner
	}

	if req.Title != "" {
		d.Title = req.Title
	}
	if req.Type != "" {
		if req.Type != "dashboard" && req.Type != "screen" {
			return fmt.Errorf("invalid dashboard type: %s", req.Type)
		}
		d.Type = req.Type
	}
	if req.Config != "" {
		d.Config = req.Config
	}
	if req.Status != nil {
		d.Status = *req.Status
	}

	return s.repo.Update(ctx, d)
}

func (s *DashboardService) Delete(ctx context.Context, id uint64, userID uint64) error {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if d.CreatedBy != userID {
		return ErrNotOwner
	}
	return s.repo.Delete(ctx, id)
}

func (s *DashboardService) EnableShare(ctx context.Context, id uint64, userID uint64) (string, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if d.CreatedBy != userID {
		return "", ErrNotOwner
	}

	links, err := s.shareLinkRepo.ListByResource(ctx, "dashboard", id)
	if err != nil {
		return "", err
	}
	for _, link := range links {
		if err := s.shareLinkRepo.Delete(ctx, link.ID); err != nil {
			return "", err
		}
	}

	link := &model.ShareLink{
		Token:        uuid.New().String(),
		ResourceType: "dashboard",
		ResourceID:   id,
		CreatedBy:    userID,
		Status:       1,
	}
	if err := s.shareLinkRepo.Create(ctx, link); err != nil {
		return "", err
	}
	return link.Token, nil
}

func (s *DashboardService) DisableShare(ctx context.Context, id uint64, userID uint64) error {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if d.CreatedBy != userID {
		return ErrNotOwner
	}

	links, err := s.shareLinkRepo.ListByResource(ctx, "dashboard", id)
	if err != nil {
		return err
	}
	for _, link := range links {
		if err := s.shareLinkRepo.Delete(ctx, link.ID); err != nil {
			return err
		}
	}
	return nil
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
