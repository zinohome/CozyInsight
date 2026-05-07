package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SaveComponents 保存仪表板组件
func (s *dashboardService) SaveComponents(ctx context.Context, dashboardID string, components []*model.DashboardComponent) error {
	// 验证仪表板存在
	dashboard, err := s.repo.Get(ctx, dashboardID)
	if err != nil {
		return fmt.Errorf("dashboard not found: %w", err)
	}

	// 删除旧组件
	if err := s.componentRepo.DeleteByDashboard(ctx, dashboardID); err != nil {
		return fmt.Errorf("failed to delete old components: %w", err)
	}

	// 验证并设置组件属性
	now := time.Now().UnixMilli()
	for i, comp := range components {
		if comp.ID == "" {
			comp.ID = uuid.New().String()
		}
		comp.DashboardID = dashboardID
		comp.CreateTime = now
		comp.UpdateTime = now

		// 验证组件位置
		if comp.W <= 0 || comp.H <= 0 {
			return fmt.Errorf("component %d: invalid size (w=%d, h=%d)", i, comp.W, comp.H)
		}
	}

	// 批量创建组件
	if len(components) > 0 {
		if err := s.componentRepo.BatchCreate(ctx, components); err != nil {
			return fmt.Errorf("failed to create components: %w", err)
		}
	}

	// 更新仪表板更新时间
	dashboard.UpdateTime = now
	if err := s.repo.Update(ctx, dashboard); err != nil {
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	return nil
}

// GetComponents 获取仪表板组件
func (s *dashboardService) GetComponents(ctx context.Context, dashboardID string) ([]*model.DashboardComponent, error) {
	// 验证仪表板存在
	if _, err := s.repo.Get(ctx, dashboardID); err != nil {
		return nil, fmt.Errorf("dashboard not found: %w", err)
	}

	components, err := s.componentRepo.ListByDashboard(ctx, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get components: %w", err)
	}

	return components, nil
}

// GetDashboardWithComponents 获取仪表板及其组件
func (s *dashboardService) GetDashboardWithComponents(ctx context.Context, dashboardID string) (*DashboardWithComponents, error) {
	dashboard, err := s.repo.Get(ctx, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("dashboard not found: %w", err)
	}

	components, err := s.componentRepo.ListByDashboard(ctx, dashboardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get components: %w", err)
	}

	return &DashboardWithComponents{
		Dashboard:  dashboard,
		Components: components,
	}, nil
}

// UpdateLayout 更新仪表板布局
func (s *dashboardService) UpdateLayout(ctx context.Context, dashboardID string, layout *model.DashboardLayout) error {
	dashboard, err := s.repo.Get(ctx, dashboardID)
	if err != nil {
		return fmt.Errorf("dashboard not found: %w", err)
	}

	// 将布局转换为JSON
	canvasStyleJSON, err := json.Marshal(layout.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal canvas style: %w", err)
	}

	componentDataJSON, err := json.Marshal(layout.Components)
	if err != nil {
		return fmt.Errorf("failed to marshal component data: %w", err)
	}

	// 更新仪表板
	dashboard.CanvasStyleData = string(canvasStyleJSON)
	dashboard.ComponentData = string(componentDataJSON)
	dashboard.UpdateTime = time.Now().UnixMilli()

	if err := s.repo.Update(ctx, dashboard); err != nil {
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	// 同时更新组件表
	if len(layout.Components) > 0 {
		if err := s.SaveComponentsBatch(ctx, dashboardID, &layout.Components); err != nil {
			return fmt.Errorf("failed to save components: %w", err)
		}
	}

	return nil
}

// SaveComponents批量保存的辅助方法
func (s *dashboardService) SaveComponentsBatch(ctx context.Context, dashboardID string, comps *[]model.DashboardComponent) error {
	components := make([]*model.DashboardComponent, len(*comps))
	for i := range *comps {
		components[i] = &(*comps)[i]
	}
	return s.SaveComponents(ctx, dashboardID, components)
}
