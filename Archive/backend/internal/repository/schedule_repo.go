package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type ScheduleRepository interface {
	Create(ctx context.Context, task *model.ScheduleTask) error
	Update(ctx context.Context, task *model.ScheduleTask) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.ScheduleTask, error)
	List(ctx context.Context) ([]*model.ScheduleTask, error)
}

type scheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository() ScheduleRepository {
	return &scheduleRepository{
		db: database.DB,
	}
}

func (r *scheduleRepository) Create(ctx context.Context, task *model.ScheduleTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *scheduleRepository) Update(ctx context.Context, task *model.ScheduleTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *scheduleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.ScheduleTask{}, "id = ?", id).Error
}

func (r *scheduleRepository) Get(ctx context.Context, id string) (*model.ScheduleTask, error) {
	var task model.ScheduleTask
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *scheduleRepository) List(ctx context.Context) ([]*model.ScheduleTask, error) {
	var tasks []*model.ScheduleTask
	err := r.db.WithContext(ctx).Order("create_time DESC").Find(&tasks).Error
	return tasks, err
}
