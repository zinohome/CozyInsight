package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"cozy-insight-backend/pkg/logger"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ChartService interface {
	Create(ctx context.Context, chart *model.ChartView) error
	Update(ctx context.Context, chart *model.ChartView) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.ChartView, error)
	List(ctx context.Context, sceneId string) ([]*model.ChartView, error)
}

type chartService struct {
	repo repository.ChartRepository
}

func NewChartService(repo repository.ChartRepository) ChartService {
	return &chartService{repo: repo}
}

// Create 创建图表
func (s *chartService) Create(ctx context.Context, chart *model.ChartView) error {
	// 验证必填字段
	if chart.Name == "" {
		return fmt.Errorf("chart name is required")
	}
	if chart.TableID == "" {
		return fmt.Errorf("table id is required")
	}
	if chart.Type == "" {
		return fmt.Errorf("chart type is required")
	}

	// 生成 ID
	if chart.ID == "" {
		chart.ID = uuid.New().String()
	}

	// 设置时间戳
	chart.CreateTime = time.Now().UnixMilli()
	chart.UpdateTime = time.Now().UnixMilli()

	logger.Log.Info("creating chart",
		zap.String("chart_id", chart.ID),
		zap.String("name", chart.Name),
		zap.String("type", chart.Type),
	)

	if err := s.repo.Create(ctx, chart); err != nil {
		logger.Log.Error("failed to create chart",
			zap.Error(err),
			zap.String("chart_id", chart.ID),
		)
		return fmt.Errorf("failed to create chart: %w", err)
	}

	logger.Log.Info("chart created successfully", zap.String("chart_id", chart.ID))
	return nil
}

// Update 更新图表
func (s *chartService) Update(ctx context.Context, chart *model.ChartView) error {
	// 验证 ID
	if chart.ID == "" {
		return fmt.Errorf("chart id is required")
	}

	// 检查图表是否存在
	existing, err := s.repo.Get(ctx, chart.ID)
	if err != nil {
		logger.Log.Error("chart not found for update",
			zap.Error(err),
			zap.String("chart_id", chart.ID),
		)
		return fmt.Errorf("chart not found: %w", err)
	}

	// 更新时间戳
	chart.UpdateTime = time.Now().UnixMilli()
	chart.CreateTime = existing.CreateTime // 保留原创建时间

	logger.Log.Info("updating chart",
		zap.String("chart_id", chart.ID),
		zap.String("name", chart.Name),
	)

	if err := s.repo.Update(ctx, chart); err != nil {
		logger.Log.Error("failed to update chart",
			zap.Error(err),
			zap.String("chart_id", chart.ID),
		)
		return fmt.Errorf("failed to update chart: %w", err)
	}

	logger.Log.Info("chart updated successfully", zap.String("chart_id", chart.ID))
	return nil
}

// Delete 删除图表
func (s *chartService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("chart id is required")
	}

	logger.Log.Info("deleting chart", zap.String("chart_id", id))

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Log.Error("failed to delete chart",
			zap.Error(err),
			zap.String("chart_id", id),
		)
		return fmt.Errorf("failed to delete chart: %w", err)
	}

	logger.Log.Info("chart deleted successfully", zap.String("chart_id", id))
	return nil
}

// Get 获取图表详情
func (s *chartService) Get(ctx context.Context, id string) (*model.ChartView, error) {
	if id == "" {
		return nil, fmt.Errorf("chart id is required")
	}

	chart, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Log.Error("failed to get chart",
			zap.Error(err),
			zap.String("chart_id", id),
		)
		return nil, fmt.Errorf("failed to get chart: %w", err)
	}

	return chart, nil
}

// List 获取图表列表
func (s *chartService) List(ctx context.Context, sceneId string) ([]*model.ChartView, error) {
	logger.Log.Info("listing charts", zap.String("scene_id", sceneId))

	list, err := s.repo.List(ctx, sceneId)
	if err != nil {
		logger.Log.Error("failed to list charts",
			zap.Error(err),
			zap.String("scene_id", sceneId),
		)
		return nil, fmt.Errorf("failed to list charts: %w", err)
	}

	logger.Log.Info("charts listed successfully",
		zap.String("scene_id", sceneId),
		zap.Int("count", len(list)),
	)
	return list, nil
}
