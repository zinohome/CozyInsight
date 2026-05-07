package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type DatasetRepository interface {
	// Dataset Group
	CreateGroup(ctx context.Context, group *model.DatasetGroup) error
	UpdateGroup(ctx context.Context, group *model.DatasetGroup) error
	DeleteGroup(ctx context.Context, id string) error
	GetGroup(ctx context.Context, id string) (*model.DatasetGroup, error)
	ListGroups(ctx context.Context) ([]*model.DatasetGroup, error)

	// Dataset Table
	CreateTable(ctx context.Context, table *model.DatasetTable) error
	UpdateTable(ctx context.Context, table *model.DatasetTable) error
	GetTable(ctx context.Context, id string) (*model.DatasetTable, error)
	ListTables(ctx context.Context, groupId string) ([]*model.DatasetTable, error)

	// Dataset Fields
	SaveFields(ctx context.Context, fields []*model.DatasetTableField) error
	GetFields(ctx context.Context, tableId string) ([]*model.DatasetTableField, error)
	DeleteFieldsByTableID(ctx context.Context, tableId string) error
	BatchCreateFields(ctx context.Context, fields []*model.DatasetTableField) error
}

type datasetRepository struct{}

func NewDatasetRepository() DatasetRepository {
	return &datasetRepository{}
}

func (r *datasetRepository) CreateGroup(ctx context.Context, group *model.DatasetGroup) error {
	return database.DB.WithContext(ctx).Create(group).Error
}

func (r *datasetRepository) UpdateGroup(ctx context.Context, group *model.DatasetGroup) error {
	return database.DB.WithContext(ctx).Save(group).Error
}

func (r *datasetRepository) DeleteGroup(ctx context.Context, id string) error {
	return database.DB.WithContext(ctx).Delete(&model.DatasetGroup{}, "id = ?", id).Error
}

func (r *datasetRepository) GetGroup(ctx context.Context, id string) (*model.DatasetGroup, error) {
	var group model.DatasetGroup
	err := database.DB.WithContext(ctx).First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *datasetRepository) ListGroups(ctx context.Context) ([]*model.DatasetGroup, error) {
	var list []*model.DatasetGroup
	err := database.DB.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *datasetRepository) CreateTable(ctx context.Context, table *model.DatasetTable) error {
	return database.DB.WithContext(ctx).Create(table).Error
}

func (r *datasetRepository) UpdateTable(ctx context.Context, table *model.DatasetTable) error {
	return database.DB.WithContext(ctx).Save(table).Error
}

func (r *datasetRepository) GetTable(ctx context.Context, id string) (*model.DatasetTable, error) {
	var table model.DatasetTable
	err := database.DB.WithContext(ctx).First(&table, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *datasetRepository) ListTables(ctx context.Context, groupId string) ([]*model.DatasetTable, error) {
	var list []*model.DatasetTable
	query := database.DB.WithContext(ctx)
	if groupId != "" {
		query = query.Where("dataset_group_id = ?", groupId)
	}
	err := query.Find(&list).Error
	return list, err
}

func (r *datasetRepository) SaveFields(ctx context.Context, fields []*model.DatasetTableField) error {
	if len(fields) == 0 {
		return nil
	}
	// Transaction to replace fields
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete old fields
		if err := tx.Delete(&model.DatasetTableField{}, "dataset_table_id = ?", fields[0].DatasetTableID).Error; err != nil {
			return err
		}
		// Create new fields
		return tx.Create(fields).Error
	})
}

func (r *datasetRepository) GetFields(ctx context.Context, tableId string) ([]*model.DatasetTableField, error) {
	var list []*model.DatasetTableField
	err := database.DB.WithContext(ctx).Where("dataset_table_id = ?", tableId).Order("column_index").Find(&list).Error
	return list, err
}

func (r *datasetRepository) DeleteFieldsByTableID(ctx context.Context, tableId string) error {
	return database.DB.WithContext(ctx).Where("dataset_table_id = ?", tableId).Delete(&model.DatasetTableField{}).Error
}

func (r *datasetRepository) BatchCreateFields(ctx context.Context, fields []*model.DatasetTableField) error {
	if len(fields) == 0 {
		return nil
	}
	return database.DB.WithContext(ctx).Create(&fields).Error
}
