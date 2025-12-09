package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type OperLogRepository interface {
	Create(ctx context.Context, log *model.SysOperLog) error
	List(ctx context.Context, userID string, module string, startTime, endTime int64, page, pageSize int) ([]*model.SysOperLog, int64, error)
	Delete(ctx context.Context, beforeTime int64) error
}

type operLogRepository struct {
	db *gorm.DB
}

func NewOperLogRepository() OperLogRepository {
	return &operLogRepository{
		db: database.DB,
	}
}

func (r *operLogRepository) Create(ctx context.Context, log *model.SysOperLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *operLogRepository) List(ctx context.Context, userID string, module string, startTime, endTime int64, page, pageSize int) ([]*model.SysOperLog, int64, error) {
	var logs []*model.SysOperLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SysOperLog{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if startTime > 0 {
		query = query.Where("create_time >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("create_time <= ?", endTime)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

func (r *operLogRepository) Delete(ctx context.Context, beforeTime int64) error {
	return r.db.WithContext(ctx).Where("create_time < ?", beforeTime).Delete(&model.SysOperLog{}).Error
}
