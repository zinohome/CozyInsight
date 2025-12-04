package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"cozy-insight-backend/pkg/logger"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestMain 在所有测试运行前初始化
func TestMain(m *testing.M) {
	// 初始化 logger
	logger.InitLogger("debug")

	// 运行测试
	code := m.Run()

	// 退出
	os.Exit(code)
}

// MockDashboardRepository 是 DashboardRepository 的 mock 实现
type MockDashboardRepository struct {
	mock.Mock
}

func (m *MockDashboardRepository) Create(ctx context.Context, dashboard *model.Dashboard) error {
	args := m.Called(ctx, dashboard)
	return args.Error(0)
}

func (m *MockDashboardRepository) Update(ctx context.Context, dashboard *model.Dashboard) error {
	args := m.Called(ctx, dashboard)
	return args.Error(0)
}

func (m *MockDashboardRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDashboardRepository) Get(ctx context.Context, id string) (*model.Dashboard, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Dashboard), args.Error(1)
}

func (m *MockDashboardRepository) List(ctx context.Context, pid string) ([]*model.Dashboard, error) {
	args := m.Called(ctx, pid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Dashboard), args.Error(1)
}

// Ensure MockDashboardRepository implements DashboardRepository
var _ repository.DashboardRepository = (*MockDashboardRepository)(nil)

func TestDashboardService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		dashboard := &model.Dashboard{
			Name:     "Test Dashboard",
			NodeType: "dashboard",
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Dashboard")).Return(nil)

		err := service.Create(ctx, dashboard)

		assert.NoError(t, err)
		assert.NotEmpty(t, dashboard.ID)
		assert.Equal(t, "0", dashboard.PID) // 默认根节点
		assert.NotZero(t, dashboard.CreateTime)
		assert.NotZero(t, dashboard.UpdateTime)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MissingName", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		dashboard := &model.Dashboard{
			NodeType: "dashboard",
		}

		err := service.Create(ctx, dashboard)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("MissingNodeType", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		dashboard := &model.Dashboard{
			Name: "Test Dashboard",
		}

		err := service.Create(ctx, dashboard)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node type is required")
	})

	t.Run("InvalidNodeType", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		dashboard := &model.Dashboard{
			Name:     "Test Dashboard",
			NodeType: "invalid",
		}

		err := service.Create(ctx, dashboard)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be 'folder' or 'dashboard'")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		dashboard := &model.Dashboard{
			Name:     "Test Dashboard",
			NodeType: "dashboard",
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Dashboard")).
			Return(errors.New("database error"))

		err := service.Create(ctx, dashboard)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create dashboard")
		mockRepo.AssertExpectations(t)
	})
}

func TestDashboardService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		existingDashboard := &model.Dashboard{
			ID:         "dashboard-1",
			Name:       "Old Dashboard",
			CreateTime: 1000,
		}

		updatedDashboard := &model.Dashboard{
			ID:   "dashboard-1",
			Name: "Updated Dashboard",
		}

		mockRepo.On("Get", ctx, "dashboard-1").Return(existingDashboard, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Dashboard")).Return(nil)

		err := service.Update(ctx, updatedDashboard)

		assert.NoError(t, err)
		assert.Equal(t, int64(1000), updatedDashboard.CreateTime)
		assert.NotZero(t, updatedDashboard.UpdateTime)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MissingID", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		dashboard := &model.Dashboard{
			Name: "Test Dashboard",
		}

		err := service.Update(ctx, dashboard)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("DashboardNotFound", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		dashboard := &model.Dashboard{
			ID:   "dashboard-1",
			Name: "Test Dashboard",
		}

		mockRepo.On("Get", ctx, "dashboard-1").Return(nil, errors.New("not found"))

		err := service.Update(ctx, dashboard)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dashboard not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestDashboardService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		mockRepo.On("Delete", ctx, "dashboard-1").Return(nil)

		err := service.Delete(ctx, "dashboard-1")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MissingID", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		err := service.Delete(ctx, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		mockRepo.On("Delete", ctx, "dashboard-1").Return(errors.New("database error"))

		err := service.Delete(ctx, "dashboard-1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete dashboard")
		mockRepo.AssertExpectations(t)
	})
}

func TestDashboardService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		expectedDashboard := &model.Dashboard{
			ID:   "dashboard-1",
			Name: "Test Dashboard",
		}

		mockRepo.On("Get", ctx, "dashboard-1").Return(expectedDashboard, nil)

		dashboard, err := service.Get(ctx, "dashboard-1")

		assert.NoError(t, err)
		assert.Equal(t, expectedDashboard, dashboard)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MissingID", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		dashboard, err := service.Get(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, dashboard)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("DashboardNotFound", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		mockRepo.On("Get", ctx, "dashboard-1").Return(nil, errors.New("not found"))

		dashboard, err := service.Get(ctx, "dashboard-1")

		assert.Error(t, err)
		assert.Nil(t, dashboard)
		assert.Contains(t, err.Error(), "failed to get dashboard")
		mockRepo.AssertExpectations(t)
	})
}

func TestDashboardService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		expectedDashboards := []*model.Dashboard{
			{ID: "dashboard-1", Name: "Dashboard 1"},
			{ID: "dashboard-2", Name: "Dashboard 2"},
		}

		mockRepo.On("List", ctx, "0").Return(expectedDashboards, nil)

		dashboards, err := service.List(ctx, "0")

		assert.NoError(t, err)
		assert.Equal(t, expectedDashboards, dashboards)
		assert.Len(t, dashboards, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("EmptyList", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		mockRepo.On("List", ctx, "").Return([]*model.Dashboard{}, nil)

		dashboards, err := service.List(ctx, "")

		assert.NoError(t, err)
		assert.Empty(t, dashboards)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockRepo := new(MockDashboardRepository)
		service := NewDashboardService(mockRepo)

		mockRepo.On("List", ctx, "0").Return(nil, errors.New("database error"))

		dashboards, err := service.List(ctx, "0")

		assert.Error(t, err)
		assert.Nil(t, dashboards)
		assert.Contains(t, err.Error(), "failed to list dashboards")
		mockRepo.AssertExpectations(t)
	})
}
