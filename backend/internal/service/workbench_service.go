package service

import (
	"context"
	"fmt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/repository"
)

type WorkbenchService struct {
	repo *repository.WorkbenchRepository
}

func NewWorkbenchService(repo *repository.WorkbenchRepository) *WorkbenchService {
	return &WorkbenchService{repo: repo}
}

func (s *WorkbenchService) GetStats(ctx context.Context, userID uint64) (*dto.WorkbenchStatsResponse, error) {
	dsCount, err := s.repo.CountByCreatedBy(ctx, "datasources", userID, "")
	if err != nil {
		return nil, fmt.Errorf("get datasource count: %w", err)
	}
	datasetCount, err := s.repo.CountByCreatedBy(ctx, "datasets", userID, "")
	if err != nil {
		return nil, fmt.Errorf("get dataset count: %w", err)
	}
	chartCount, err := s.repo.CountByCreatedBy(ctx, "charts", userID, "")
	if err != nil {
		return nil, fmt.Errorf("get chart count: %w", err)
	}
	dbCount, err := s.repo.CountByCreatedBy(ctx, "dashboards", userID, "type = 'dashboard'")
	if err != nil {
		return nil, fmt.Errorf("get dashboard count: %w", err)
	}
	screenCount, err := s.repo.CountByCreatedBy(ctx, "dashboards", userID, "type = 'screen'")
	if err != nil {
		return nil, fmt.Errorf("get screen count: %w", err)
	}
	return &dto.WorkbenchStatsResponse{
		DatasourceCount: dsCount,
		DatasetCount:    datasetCount,
		ChartCount:      chartCount,
		DashboardCount:  dbCount,
		ScreenCount:     screenCount,
	}, nil
}

func (s *WorkbenchService) RecordVisit(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	return s.repo.UpsertRecentView(ctx, userID, resourceType, resourceID)
}

func (s *WorkbenchService) ListRecentViews(ctx context.Context, userID uint64) ([]dto.RecentViewItem, error) {
	return s.repo.ListRecentViews(ctx, userID, 20)
}

func (s *WorkbenchService) AddFavorite(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	return s.repo.AddFavorite(ctx, userID, resourceType, resourceID)
}

func (s *WorkbenchService) DeleteFavorite(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	return s.repo.DeleteFavorite(ctx, userID, resourceType, resourceID)
}

func (s *WorkbenchService) ListFavorites(ctx context.Context, userID uint64) ([]dto.FavoriteItem, error) {
	return s.repo.ListFavorites(ctx, userID)
}
