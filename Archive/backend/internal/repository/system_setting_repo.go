package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type SystemSettingRepository interface {
	Create(ctx context.Context, setting *model.SysSetting) error
	Update(ctx context.Context, setting *model.SysSetting) error
	Delete(ctx context.Context, key string) error
	GetByKey(ctx context.Context, key string) (*model.SysSetting, error)
	ListByType(ctx context.Context, settingType string) ([]*model.SysSetting, error)
}

type systemSettingRepository struct {
	db *gorm.DB
}

func NewSystemSettingRepository() SystemSettingRepository {
	return &systemSettingRepository{
		db: database.DB,
	}
}

func (r *systemSettingRepository) Create(ctx context.Context, setting *model.SysSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *systemSettingRepository) Update(ctx context.Context, setting *model.SysSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

func (r *systemSettingRepository) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("setting_key = ?", key).Delete(&model.SysSetting{}).Error
}

func (r *systemSettingRepository) GetByKey(ctx context.Context, key string) (*model.SysSetting, error) {
	var setting model.SysSetting
	err := r.db.WithContext(ctx).Where("setting_key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *systemSettingRepository) ListByType(ctx context.Context, settingType string) ([]*model.SysSetting, error) {
	var settings []*model.SysSetting
	err := r.db.WithContext(ctx).Where("type = ?", settingType).Find(&settings).Error
	return settings, err
}
