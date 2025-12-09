package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ShareService interface {
	CreateShare(ctx context.Context, share *model.Share) error
	GetShare(ctx context.Context, id string) (*model.Share, error)
	GetShareByToken(ctx context.Context, token string) (*model.Share, error)
	DeleteShare(ctx context.Context, id string) error
	ListShares(ctx context.Context, resourceType, resourceID string) ([]*model.Share, error)
	ValidateShare(ctx context.Context, token, password string) (*model.Share, error)
}

type shareService struct {
	repo repository.ShareRepository
}

func NewShareService(repo repository.ShareRepository) ShareService {
	return &shareService{repo: repo}
}

func (s *shareService) CreateShare(ctx context.Context, share *model.Share) error {
	if share.ResourceType == "" || share.ResourceID == "" {
		return fmt.Errorf("resource type and id are required")
	}

	if share.ID == "" {
		share.ID = uuid.New().String()
	}
	
	// 生成分享token
	if share.Token == "" {
		share.Token = generateToken()
	}
	
	share.CreateTime = time.Now().UnixMilli()
	
	// 设置过期时间(默认7天)
	if share.ExpireTime == 0 {
		share.ExpireTime = time.Now().Add(7 * 24 * time.Hour).UnixMilli()
	}

	return s.repo.Create(ctx, share)
}

func (s *shareService) GetShare(ctx context.Context, id string) (*model.Share, error) {
	return s.repo.Get(ctx, id)
}

func (s *shareService) GetShareByToken(ctx context.Context, token string) (*model.Share, error) {
	return s.repo.GetByToken(ctx, token)
}

func (s *shareService) DeleteShare(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *shareService) ListShares(ctx context.Context, resourceType, resourceID string) ([]*model.Share, error) {
	return s.repo.List(ctx, resourceType, resourceID)
}

func (s *shareService) ValidateShare(ctx context.Context, token, password string) (*model.Share, error) {
	share, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("share not found")
	}

	// 检查是否过期
	if share.ExpireTime > 0 && time.Now().UnixMilli() > share.ExpireTime {
		return nil, fmt.Errorf("share expired")
	}

	// 检查密码
	if share.Password != "" && share.Password != password {
		return nil, fmt.Errorf("invalid password")
	}

	return share, nil
}

// generateToken 生成分享token
func generateToken() string {
	return uuid.New().String()[:8]
}
