package service

import (
	"context"
	"encoding/json"
	"fmt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type ChartService struct {
	repo *repository.ChartRepository
}

func NewChartService(repo *repository.ChartRepository) *ChartService {
	return &ChartService{repo: repo}
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
	if req.DatasetID != 0 {
		chart.DatasetID = req.DatasetID
	}
	if req.Config != "" {
		chart.Config = req.Config
	}
	chart.Status = req.Status

	return s.repo.Update(ctx, chart)
}

func (s *ChartService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
