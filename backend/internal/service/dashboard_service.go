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

type DashboardService interface {
	Create(ctx context.Context, dashboard *model.Dashboard) error
	Update(ctx context.Context, dashboard *model.Dashboard) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Dashboard, error)
	List(ctx context.Context, pid string) ([]*model.Dashboard, error)

	// 发布管理
	Publish(ctx context.Context, id string, userId string) error
	Unpublish(ctx context.Context, id string) error
	GetPublishStatus(ctx context.Context, id string) (int, error)
}

type dashboardService struct {
	repo repository.DashboardRepository
}

func NewDashboardService(repo repository.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo}
}

// Create 创建仪表板
func (s *dashboardService) Create(ctx context.Context, dashboard *model.Dashboard) error {
	// 验证必填字段
	if dashboard.Name == "" {
		return fmt.Errorf("dashboard name is required")
	}
	if dashboard.NodeType == "" {
		return fmt.Errorf("node type is required")
	}
	// NodeType 只能是 folder 或 dashboard
	if dashboard.NodeType != "folder" && dashboard.NodeType != "dashboard" {
		return fmt.Errorf("node type must be 'folder' or 'dashboard'")
	}

	// 生成 ID
	if dashboard.ID == "" {
		dashboard.ID = uuid.New().String()
	}

	// 设置默认值
	if dashboard.PID == "" {
		dashboard.PID = "0" // 默认根节点
	}
	if dashboard.Status == 0 {
		dashboard.Status = 0 // 未发布
	}

	// 设置时间戳
	dashboard.CreateTime = time.Now().UnixMilli()
	dashboard.UpdateTime = time.Now().UnixMilli()

	logger.Log.Info("creating dashboard",
		zap.String("dashboard_id", dashboard.ID),
		zap.String("name", dashboard.Name),
		zap.String("node_type", dashboard.NodeType),
	)

	if err := s.repo.Create(ctx, dashboard); err != nil {
		logger.Log.Error("failed to create dashboard",
			zap.Error(err),
			zap.String("dashboard_id", dashboard.ID),
		)
		return fmt.Errorf("failed to create dashboard: %w", err)
	}

	logger.Log.Info("dashboard created successfully", zap.String("dashboard_id", dashboard.ID))
	return nil
}

// Update 更新仪表板
func (s *dashboardService) Update(ctx context.Context, dashboard *model.Dashboard) error {
	// 验证 ID
	if dashboard.ID == "" {
		return fmt.Errorf("dashboard id is required")
	}

	// 检查是否存在
	existing, err := s.repo.Get(ctx, dashboard.ID)
	if err != nil {
		logger.Log.Error("dashboard not found for update",
			zap.Error(err),
			zap.String("dashboard_id", dashboard.ID),
		)
		return fmt.Errorf("dashboard not found: %w", err)
	}

	// 更新时间戳
	dashboard.UpdateTime = time.Now().UnixMilli()
	dashboard.CreateTime = existing.CreateTime // 保留原创建时间

	logger.Log.Info("updating dashboard",
		zap.String("dashboard_id", dashboard.ID),
		zap.String("name", dashboard.Name),
	)

	if err := s.repo.Update(ctx, dashboard); err != nil {
		logger.Log.Error("failed to update dashboard",
			zap.Error(err),
			zap.String("dashboard_id", dashboard.ID),
		)
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	logger.Log.Info("dashboard updated successfully", zap.String("dashboard_id", dashboard.ID))
	return nil
}

// Delete 删除仪表板
func (s *dashboardService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("dashboard id is required")
	}

	logger.Log.Info("deleting dashboard", zap.String("dashboard_id", id))

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Log.Error("failed to delete dashboard",
			zap.Error(err),
			zap.String("dashboard_id", id),
		)
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	logger.Log.Info("dashboard deleted successfully", zap.String("dashboard_id", id))
	return nil
}

// Get 获取仪表板详情
func (s *dashboardService) Get(ctx context.Context, id string) (*model.Dashboard, error) {
	if id == "" {
		return nil, fmt.Errorf("dashboard id is required")
	}

	dashboard, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Log.Error("failed to get dashboard",
			zap.Error(err),
			zap.String("dashboard_id", id),
		)
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	return dashboard, nil
}

// List 获取仪表板列表
func (s *dashboardService) List(ctx context.Context, pid string) ([]*model.Dashboard, error) {
	logger.Log.Info("listing dashboards", zap.String("pid", pid))

	list, err := s.repo.List(ctx, pid)
	if err != nil {
		logger.Log.Error("failed to list dashboards",
			zap.Error(err),
			zap.String("pid", pid),
		)
		return nil, fmt.Errorf("failed to list dashboards: %w", err)
	}

	logger.Log.Info("dashboards listed successfully",
		zap.String("pid", pid),
		zap.Int("count", len(list)),
	)
	return list, nil
}

// Publish 发布仪表板
func (s *dashboardService) Publish(ctx context.Context, id string, userId string) error {
	dashboard, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	dashboard.Status = 1 // 1=已发布
	dashboard.UpdateTime = time.Now().UnixMilli()
	if userId != "" {
		dashboard.CreateBy = userId // 记录发布人
	}

	return s.repo.Update(ctx, dashboard)
}

// Unpublish 下线仪表板
func (s *dashboardService) Unpublish(ctx context.Context, id string) error {
	dashboard, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	dashboard.Status = 0 // 0=未发布
	dashboard.UpdateTime = time.Now().UnixMilli()

	return s.repo.Update(ctx, dashboard)
}

// GetPublishStatus 获取发布状态
func (s *dashboardService) GetPublishStatus(ctx context.Context, id string) (int, error) {
	dashboard, err := s.repo.Get(ctx, id)
	if err != nil {
		return 0, err
	}
	return dashboard.Status, nil
}
