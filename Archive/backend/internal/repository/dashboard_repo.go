package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"
)

type DashboardRepository interface {
	Create(ctx context.Context, dashboard *model.Dashboard) error
	Update(ctx context.Context, dashboard *model.Dashboard) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Dashboard, error)
	List(ctx context.Context, pid string) ([]*model.Dashboard, error)
}

type dashboardRepository struct{}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{}
}

func (r *dashboardRepository) Create(ctx context.Context, dashboard *model.Dashboard) error {
	return database.DB.WithContext(ctx).Create(dashboard).Error
}

func (r *dashboardRepository) Update(ctx context.Context, dashboard *model.Dashboard) error {
	return database.DB.WithContext(ctx).Save(dashboard).Error
}

func (r *dashboardRepository) Delete(ctx context.Context, id string) error {
	return database.DB.WithContext(ctx).Delete(&model.Dashboard{}, "id = ?", id).Error
}

func (r *dashboardRepository) Get(ctx context.Context, id string) (*model.Dashboard, error) {
	var dashboard model.Dashboard
	err := database.DB.WithContext(ctx).First(&dashboard, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &dashboard, nil
}

func (r *dashboardRepository) List(ctx context.Context, pid string) ([]*model.Dashboard, error) {
	var list []*model.Dashboard
	query := database.DB.WithContext(ctx)
	if pid != "" {
		query = query.Where("pid = ?", pid)
	}
	err := query.Order("sort asc, create_time desc").Find(&list).Error
	return list, err
}
