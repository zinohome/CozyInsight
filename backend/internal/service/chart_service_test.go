package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChartRepository 是 ChartRepository 的 mock 实现
type MockChartRepository struct {
	mock.Mock
}

func (m *MockChartRepository) Create(ctx context.Context, chart *model.ChartView) error {
	args := m.Called(ctx, chart)
	return args.Error(0)
}

func (m *MockChartRepository) Update(ctx context.Context, chart *model.ChartView) error {
	args := m.Called(ctx, chart)
	return args.Error(0)
}

func (m *MockChartRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChartRepository) Get(ctx context.Context, id string) (*model.ChartView, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ChartView), args.Error(1)
}

func (m *MockChartRepository) List(ctx context.Context, sceneId string) ([]*model.ChartView, error) {
	args := m.Called(ctx, sceneId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ChartView), args.Error(1)
}

// Ensure MockChartRepository implements ChartRepository
var _ repository.ChartRepository = (*MockChartRepository)(nil)

func TestChartService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		chart := &model.ChartView{
			Name:    "Test Chart",
			TableID: "table-1",
			Type:    "bar",
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.ChartView")).Return(nil)

		err := service.Create(ctx, chart)

		assert.NoError(t, err)
		assert.NotEmpty(t, chart.ID)
		assert.NotZero(t, chart.CreateTime)
		assert.NotZero(t, chart.UpdateTime)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MissingName", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		chart := &model.ChartView{
			TableID: "table-1",
			Type:    "bar",
		}

		err := service.Create(ctx, chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("MissingTableID", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		chart := &model.ChartView{
			Name: "Test Chart",
			Type: "bar",
		}

		err := service.Create(ctx, chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table id is required")
	})

	t.Run("MissingType", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		chart := &model.ChartView{
			Name:    "Test Chart",
			TableID: "table-1",
		}

		err := service.Create(ctx, chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type is required")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		chart := &model.ChartView{
			Name:    "Test Chart",
			TableID: "table-1",
			Type:    "bar",
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.ChartView")).
			Return(errors.New("database error"))

		err := service.Create(ctx, chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create chart")
		mockRepo.AssertExpectations(t)
	})
}

func TestChartService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		existingChart := &model.ChartView{
			ID:         "chart-1",
			Name:       "Old Chart",
			CreateTime: 1000,
		}

		updatedChart := &model.ChartView{
			ID:   "chart-1",
			Name: "Updated Chart",
		}

		mockRepo.On("Get", ctx, "chart-1").Return(existingChart, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.ChartView")).Return(nil)

		err := service.Update(ctx, updatedChart)

		assert.NoError(t, err)
		assert.Equal(t, int64(1000), updatedChart.CreateTime)
		assert.NotZero(t, updatedChart.UpdateTime)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MissingID", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		chart := &model.ChartView{
			Name: "Test Chart",
		}

		err := service.Update(ctx, chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("ChartNotFound", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		chart := &model.ChartView{
			ID:   "chart-1",
			Name: "Test Chart",
		}

		mockRepo.On("Get", ctx, "chart-1").Return(nil, errors.New("not found"))

		err := service.Update(ctx, chart)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chart not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestChartService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		mockRepo.On("Delete", ctx, "chart-1").Return(nil)

		err := service.Delete(ctx, "chart-1")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MissingID", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		err := service.Delete(ctx, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		mockRepo.On("Delete", ctx, "chart-1").Return(errors.New("database error"))

		err := service.Delete(ctx, "chart-1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete chart")
		mockRepo.AssertExpectations(t)
	})
}

func TestChartService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		expectedChart := &model.ChartView{
			ID:   "chart-1",
			Name: "Test Chart",
		}

		mockRepo.On("Get", ctx, "chart-1").Return(expectedChart, nil)

		chart, err := service.Get(ctx, "chart-1")

		assert.NoError(t, err)
		assert.Equal(t, expectedChart, chart)
		mockRepo.AssertExpectations(t)
	})

	t.Run("MissingID", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		chart, err := service.Get(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, chart)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("ChartNotFound", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		mockRepo.On("Get", ctx, "chart-1").Return(nil, errors.New("not found"))

		chart, err := service.Get(ctx, "chart-1")

		assert.Error(t, err)
		assert.Nil(t, chart)
		assert.Contains(t, err.Error(), "failed to get chart")
		mockRepo.AssertExpectations(t)
	})
}

func TestChartService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		expectedCharts := []*model.ChartView{
			{ID: "chart-1", Name: "Chart 1"},
			{ID: "chart-2", Name: "Chart 2"},
		}

		mockRepo.On("List", ctx, "scene-1").Return(expectedCharts, nil)

		charts, err := service.List(ctx, "scene-1")

		assert.NoError(t, err)
		assert.Equal(t, expectedCharts, charts)
		assert.Len(t, charts, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("EmptyList", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		mockRepo.On("List", ctx, "").Return([]*model.ChartView{}, nil)

		charts, err := service.List(ctx, "")

		assert.NoError(t, err)
		assert.Empty(t, charts)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		mockRepo := new(MockChartRepository)
		service := NewChartService(mockRepo)

		mockRepo.On("List", ctx, "scene-1").Return(nil, errors.New("database error"))

		charts, err := service.List(ctx, "scene-1")

		assert.Error(t, err)
		assert.Nil(t, charts)
		assert.Contains(t, err.Error(), "failed to list charts")
		mockRepo.AssertExpectations(t)
	})
}
