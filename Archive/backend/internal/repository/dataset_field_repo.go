package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type DatasetFieldRepository interface {
	Create(ctx context.Context, field *model.DatasetTableField) error
	Update(ctx context.Context, field *model.DatasetTableField) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*model.DatasetTableField, error)
	ListByTableID(ctx context.Context, tableId string) ([]*model.DatasetTableField, error)
	BatchCreate(ctx context.Context, fields []*model.DatasetTableField) error
}

type datasetFieldRepository struct {
	db *gorm.DB
}

func NewDatasetFieldRepository() DatasetFieldRepository {
	return &datasetFieldRepository{
		db: database.DB,
	}
}

func (r *datasetFieldRepository) Create(ctx context.Context, field *model.DatasetTableField) error {
	return r.db.WithContext(ctx).Create(field).Error
}

func (r *datasetFieldRepository) Update(ctx context.Context, field *model.DatasetTableField) error {
	return r.db.WithContext(ctx).Save(field).Error
}

func (r *datasetFieldRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.DatasetTableField{}, "id = ?", id).Error
}

func (r *datasetFieldRepository) GetByID(ctx context.Context, id string) (*model.DatasetTableField, error) {
	var field model.DatasetTableField
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&field).Error
	if err != nil {
		return nil, err
	}
	return &field, nil
}

func (r *datasetFieldRepository) ListByTableID(ctx context.Context, tableId string) ([]*model.DatasetTableField, error) {
	var fields []*model.DatasetTableField
	err := r.db.WithContext(ctx).
		Where("table_id = ?", tableId).
		Order("sort_index ASC, create_time ASC").
		Find(&fields).Error
	return fields, err
}

func (r *datasetFieldRepository) BatchCreate(ctx context.Context, fields []*model.DatasetTableField) error {
	return r.db.WithContext(ctx).Create(&fields).Error
}
