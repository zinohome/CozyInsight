package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type RowPermissionRepository interface {
	Create(ctx context.Context, permission *model.DatasetRowPermissions) error
	Update(ctx context.Context, permission *model.DatasetRowPermissions) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.DatasetRowPermissions, error)
	ListByDataset(ctx context.Context, datasetID string) ([]*model.DatasetRowPermissions, error)
}

type rowPermissionRepository struct {
	db *gorm.DB
}

func NewRowPermissionRepository() RowPermissionRepository {
	return &rowPermissionRepository{
		db: database.DB,
	}
}

func (r *rowPermissionRepository) Create(ctx context.Context, permission *model.DatasetRowPermissions) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *rowPermissionRepository) Update(ctx context.Context, permission *model.DatasetRowPermissions) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *rowPermissionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.DatasetRowPermissions{}, "id = ?", id).Error
}

func (r *rowPermissionRepository) Get(ctx context.Context, id string) (*model.DatasetRowPermissions, error) {
	var permission model.DatasetRowPermissions
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *rowPermissionRepository) ListByDataset(ctx context.Context, datasetID string) ([]*model.DatasetRowPermissions, error) {
	var permissions []*model.DatasetRowPermissions
	err := r.db.WithContext(ctx).
		Where("dataset_id = ? AND enable = ?", datasetID, true).
		Find(&permissions).Error
	return permissions, err
}
