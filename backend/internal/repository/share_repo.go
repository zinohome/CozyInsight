package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type ShareRepository interface {
	Create(ctx context.Context, share *model.Share) error
	Get(ctx context.Context, id string) (*model.Share, error)
	GetByToken(ctx context.Context, token string) (*model.Share, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, resourceType, resourceID string) ([]*model.Share, error)
}

type shareRepository struct {
	db *gorm.DB
}

func NewShareRepository() ShareRepository {
	return &shareRepository{
		db: database.DB,
	}
}

func (r *shareRepository) Create(ctx context.Context, share *model.Share) error {
	return r.db.WithContext(ctx).Create(share).Error
}

func (r *shareRepository) Get(ctx context.Context, id string) (*model.Share, error) {
	var share model.Share
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *shareRepository) GetByToken(ctx context.Context, token string) (*model.Share, error) {
	var share model.Share
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *shareRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Share{}, "id = ?", id).Error
}

func (r *shareRepository) List(ctx context.Context, resourceType, resourceID string) ([]*model.Share, error) {
	var shares []*model.Share
	query := r.db.WithContext(ctx)
	
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if resourceID != "" {
		query = query.Where("resource_id = ?", resourceID)
	}
	
	err := query.Order("create_time DESC").Find(&shares).Error
	return shares, err
}
