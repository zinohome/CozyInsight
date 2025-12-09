package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type SystemSettingService interface {
	Get(ctx context.Context, key string) (*model.SysSetting, error)
	Set(ctx context.Context, settingType, key, value string, updateBy string) error
	GetByType(ctx context.Context, settingType string) ([]*model.SysSetting, error)
	Delete(ctx context.Context, key string) error
}

type systemSettingService struct {
	repo repository.SystemSettingRepository
}

func NewSystemSettingService(repo repository.SystemSettingRepository) SystemSettingService {
	return &systemSettingService{repo: repo}
}

func (s *systemSettingService) Get(ctx context.Context, key string) (*model.SysSetting, error) {
	return s.repo.GetByKey(ctx, key)
}

func (s *systemSettingService) Set(ctx context.Context, settingType, key, value string, updateBy string) error {
	if key == "" {
		return fmt.Errorf("setting key is required")
	}

	setting, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		// 不存在,创建新的
		setting = &model.SysSetting{
			ID:         uuid.New().String(),
			Type:       settingType,
			SettingKey: key,
			Value:      value,
			UpdateBy:   updateBy,
		}
		return s.repo.Create(ctx, setting)
	}

	// 更新已有配置
	setting.Value = value
	setting.UpdateBy = updateBy
	return s.repo.Update(ctx, setting)
}

func (s *systemSettingService) GetByType(ctx context.Context, settingType string) ([]*model.SysSetting, error) {
	return s.repo.ListByType(ctx, settingType)
}

func (s *systemSettingService) Delete(ctx context.Context, key string) error {
	return s.repo.Delete(ctx, key)
}
