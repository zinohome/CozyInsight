package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type DashboardComponentRepository interface {
	Create(ctx context.Context, component *model.DashboardComponent) error
	Update(ctx context.Context, component *model.DashboardComponent) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.DashboardComponent, error)
	ListByDashboard(ctx context.Context, dashboardID string) ([]*model.DashboardComponent, error)
	BatchCreate(ctx context.Context, components []*model.DashboardComponent) error
	DeleteByDashboard(ctx context.Context, dashboardID string) error
}

type dashboardComponentRepository struct {
	db *gorm.DB
}

func NewDashboardComponentRepository() DashboardComponentRepository {
	return &dashboardComponentRepository{
		db: database.DB,
	}
}

func (r *dashboardComponentRepository) Create(ctx context.Context, component *model.DashboardComponent) error {
	return r.db.WithContext(ctx).Create(component).Error
}

func (r *dashboardComponentRepository) Update(ctx context.Context, component *model.DashboardComponent) error {
	return r.db.WithContext(ctx).Save(component).Error
}

func (r *dashboardComponentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.DashboardComponent{}, "id = ?", id).Error
}

func (r *dashboardComponentRepository) Get(ctx context.Context, id string) (*model.DashboardComponent, error) {
	var component model.DashboardComponent
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&component).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

func (r *dashboardComponentRepository) ListByDashboard(ctx context.Context, dashboardID string) ([]*model.DashboardComponent, error) {
	var components []*model.DashboardComponent
	err := r.db.WithContext(ctx).
		Where("dashboard_id = ?", dashboardID).
		Order("y ASC, x ASC"). // 按位置排序
		Find(&components).Error
	return components, err
}

func (r *dashboardComponentRepository) BatchCreate(ctx context.Context, components []*model.DashboardComponent) error {
	if len(components) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&components).Error
}

func (r *dashboardComponentRepository) DeleteByDashboard(ctx context.Context, dashboardID string) error {
	return r.db.WithContext(ctx).
		Where("dashboard_id = ?", dashboardID).
		Delete(&model.DashboardComponent{}).Error
}
