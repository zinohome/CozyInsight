package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"
)

type DatasourceRepository interface {
	Create(ctx context.Context, ds *model.Datasource) error
	Update(ctx context.Context, ds *model.Datasource) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*model.Datasource, error)
	List(ctx context.Context) ([]*model.Datasource, error)
}

type datasourceRepository struct{}

func NewDatasourceRepository() DatasourceRepository {
	return &datasourceRepository{}
}

func (r *datasourceRepository) Create(ctx context.Context, ds *model.Datasource) error {
	return database.DB.WithContext(ctx).Create(ds).Error
}

func (r *datasourceRepository) Update(ctx context.Context, ds *model.Datasource) error {
	return database.DB.WithContext(ctx).Save(ds).Error
}

func (r *datasourceRepository) Delete(ctx context.Context, id string) error {
	return database.DB.WithContext(ctx).Delete(&model.Datasource{}, "id = ?", id).Error
}

func (r *datasourceRepository) GetByID(ctx context.Context, id string) (*model.Datasource, error) {
	var ds model.Datasource
	err := database.DB.WithContext(ctx).First(&ds, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *datasourceRepository) List(ctx context.Context) ([]*model.Datasource, error) {
	var list []*model.Datasource
	err := database.DB.WithContext(ctx).Find(&list).Error
	return list, err
}
