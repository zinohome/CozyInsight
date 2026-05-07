package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"
)

type ChartRepository interface {
	Create(ctx context.Context, chart *model.ChartView) error
	Update(ctx context.Context, chart *model.ChartView) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.ChartView, error)
	List(ctx context.Context, sceneId string) ([]*model.ChartView, error)
}

type chartRepository struct{}

func NewChartRepository() ChartRepository {
	return &chartRepository{}
}

func (r *chartRepository) Create(ctx context.Context, chart *model.ChartView) error {
	return database.DB.WithContext(ctx).Create(chart).Error
}

func (r *chartRepository) Update(ctx context.Context, chart *model.ChartView) error {
	return database.DB.WithContext(ctx).Save(chart).Error
}

func (r *chartRepository) Delete(ctx context.Context, id string) error {
	return database.DB.WithContext(ctx).Delete(&model.ChartView{}, "id = ?", id).Error
}

func (r *chartRepository) Get(ctx context.Context, id string) (*model.ChartView, error) {
	var chart model.ChartView
	err := database.DB.WithContext(ctx).First(&chart, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &chart, nil
}

func (r *chartRepository) List(ctx context.Context, sceneId string) ([]*model.ChartView, error) {
	var list []*model.ChartView
	query := database.DB.WithContext(ctx)
	if sceneId != "" {
		query = query.Where("scene_id = ?", sceneId)
	}
	err := query.Order("create_time desc").Find(&list).Error
	return list, err
}
