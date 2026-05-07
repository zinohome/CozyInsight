package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type CalculatedFieldRepository interface {
	Create(ctx context.Context, field *model.DatasetTableFieldCalculated) error
	Update(ctx context.Context, field *model.DatasetTableFieldCalculated) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.DatasetTableFieldCalculated, error)
	ListByTable(ctx context.Context, tableID string) ([]*model.DatasetTableFieldCalculated, error)
}

type calculatedFieldRepository struct {
	db *gorm.DB
}

func NewCalculatedFieldRepository() CalculatedFieldRepository {
	return &calculatedFieldRepository{
		db: database.DB,
	}
}

func (r *calculatedFieldRepository) Create(ctx context.Context, field *model.DatasetTableFieldCalculated) error {
	return r.db.WithContext(ctx).Create(field).Error
}

func (r *calculatedFieldRepository) Update(ctx context.Context, field *model.DatasetTableFieldCalculated) error {
	return r.db.WithContext(ctx).Save(field).Error
}

func (r *calculatedFieldRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.DatasetTableFieldCalculated{}, "id = ?", id).Error
}

func (r *calculatedFieldRepository) Get(ctx context.Context, id string) (*model.DatasetTableFieldCalculated, error) {
	var field model.DatasetTableFieldCalculated
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&field).Error
	if err != nil {
		return nil, err
	}
	return &field, nil
}

func (r *calculatedFieldRepository) ListByTable(ctx context.Context, tableID string) ([]*model.DatasetTableFieldCalculated, error) {
	var fields []*model.DatasetTableFieldCalculated
	err := r.db.WithContext(ctx).Where("dataset_table_id = ?", tableID).Find(&fields).Error
	return fields, err
}
