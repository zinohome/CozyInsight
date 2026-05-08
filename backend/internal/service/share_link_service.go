package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type ShareLinkService struct {
	repo          *repository.ShareLinkRepository
	dashboardRepo *repository.DashboardRepository
}

func NewShareLinkService(repo *repository.ShareLinkRepository, dashboardRepo *repository.DashboardRepository) *ShareLinkService {
	return &ShareLinkService{repo: repo, dashboardRepo: dashboardRepo}
}

func (s *ShareLinkService) Create(ctx context.Context, resourceType string, resourceID uint64, userID uint64) (*model.ShareLink, error) {
	token := uuid.New().String()
	link := &model.ShareLink{
		Token:        token,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedBy:    userID,
		Status:       1,
	}
	if err := s.repo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("create share link failed: %w", err)
	}
	return link, nil
}

func (s *ShareLinkService) GetDashboard(ctx context.Context, token string) (*model.Dashboard, error) {
	link, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("share link not found: %w", err)
	}
	if link.Status != 1 {
		return nil, fmt.Errorf("share link disabled")
	}
	if link.ExpireAt != nil && link.ExpireAt.Before(time.Now()) {
		return nil, fmt.Errorf("share link expired")
	}
	if link.ResourceType != "dashboard" {
		return nil, fmt.Errorf("invalid resource type")
	}
	dashboard, err := s.dashboardRepo.FindByID(ctx, link.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("dashboard not found: %w", err)
	}
	return dashboard, nil
}

func (s *ShareLinkService) Revoke(ctx context.Context, token string) error {
	link, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("share link not found: %w", err)
	}
	return s.repo.Delete(ctx, link.ID)
}
