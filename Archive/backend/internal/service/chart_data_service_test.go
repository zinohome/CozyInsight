package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
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

func (m *MockChartRepository) List(ctx context.Context, sceneID string) ([]*model.ChartView, error) {
	args := m.Called(ctx, sceneID)
	return args.Get(0).([]*model.ChartView), args.Error(1)
}

// Mock Dataset Repository
type MockDatasetRepository struct {
	mock.Mock
}

func (m *MockDatasetRepository) CreateGroup(ctx context.Context, group *model.DatasetGroup) error {
	return nil
}

func (m *MockDatasetRepository) ListGroups(ctx context.Context) ([]*model.DatasetGroup, error) {
	return nil, nil
}

func (m *MockDatasetRepository) CreateTable(ctx context.Context, table *model.DatasetTable) error {
	return nil
}

func (m *MockDatasetRepository) UpdateTable(ctx context.Context, table *model.DatasetTable) error {
	return nil
}

func (m *MockDatasetRepository) GetTable(ctx context.Context, id string) (*model.DatasetTable, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DatasetTable), args.Error(1)
}

func (m *MockDatasetRepository) ListTables(ctx context.Context, groupId string) ([]*model.DatasetTable, error) {
	return nil, nil
}

func (m *MockDatasetRepository) SaveFields(ctx context.Context, fields []*model.DatasetTableField) error {
	return nil
}

func (m *MockDatasetRepository) GetFields(ctx context.Context, tableId string) ([]*model.DatasetTableField, error) {
	return nil, nil
}

func (m *MockDatasetRepository) DeleteFieldsByTableID(ctx context.Context, tableId string) error {
	return nil
}

func (m *MockDatasetRepository) BatchCreateFields(ctx context.Context, fields []*model.DatasetTableField) error {
	return nil
}

// Test buildChartSQL
func TestBuildChartSQL_WithAggregates(t *testing.T) {
	service := &chartDataService{}

	chart := &model.ChartView{
		ID:      "chart1",
		TableID: "table1",
		Type:    "bar",
		XAxis:   `{\"fields\":[{\"name\":\"category\"}]}`,
		YAxis:   `{\"fields\":[{\"name\":\"amount\",\"aggregate\":\"SUM\",\"sort\":\"DESC\"}]}`,
	}

	dataset := &model.DatasetTable{
		ID:                "table1",
		PhysicalTableName: "sales",
		Type:              "db",
	}

	sql, err := service.buildChartSQL(chart, dataset, nil)
	assert.NoError(t, err)
	assert.Contains(t, sql, "SELECT")
	assert.Contains(t, sql, "category")
	assert.Contains(t, sql, "SUM(amount)")
	assert.Contains(t, sql, "GROUP BY category")
	assert.Contains(t, sql, "ORDER BY amount DESC")
}

func TestBuildChartSQL_WithFilter(t *testing.T) {
	service := &chartDataService{}

	chart := &model.ChartView{
		ID:      "chart1",
		TableID: "table1",
		Type:    "line",
		XAxis:   `{\"fields\":[{\"name\":\"date\"}]}`,
		YAxis:   `{\"fields\":[{\"name\":\"revenue\",\"aggregate\":\"SUM\"}]}`,
	}

	dataset := &model.DatasetTable{
		ID:                "table1",
		PhysicalTableName: "orders",
		Type:              "db",
	}

	filter := &QueryFilter{
		Filters: []FilterCondition{
			{Field: "status", Operator: "=", Value: "completed"},
			{Field: "amount", Operator: ">", Value: 100},
		},
		Limit:  50,
		Offset: 0,
	}

	sql, err := service.buildChartSQL(chart, dataset, filter)
	assert.NoError(t, err)
	assert.Contains(t, sql, "WHERE")
	assert.Contains(t, sql, "LIMIT 50")
}

func TestBuildChartSQL_MultipleAggregates(t *testing.T) {
	service := &chartDataService{}

	chart := &model.ChartView{
		ID:      "chart1",
		TableID: "table1",
		Type:    "bar",
		XAxis:   `{\"fields\":[{\"name\":\"category\"}]}`,
		YAxis: `{\"fields\":[
			{\"name\":\"total\",\"aggregate\":\"SUM\"},
			{\"name\":\"average\",\"aggregate\":\"AVG\"},
			{\"name\":\"count\",\"aggregate\":\"COUNT\"}
		]}`,
	}

	dataset := &model.DatasetTable{
		ID:                "table1",
		PhysicalTableName: "sales",
		Type:              "db",
	}

	sql, err := service.buildChartSQL(chart, dataset, nil)
	assert.NoError(t, err)
	assert.Contains(t, sql, "SUM(total)")
	assert.Contains(t, sql, "AVG(average)")
	assert.Contains(t, sql, "COUNT(count)")
}

func TestBuildWhereConditions(t *testing.T) {
	service := &chartDataService{}

	tests := []struct {
		name     string
		filters  []FilterCondition
		expected int // 期望的条件数量
	}{
		{
			name: "等于条件",
			filters: []FilterCondition{
				{Field: "status", Operator: "=", Value: "active"},
			},
			expected: 1,
		},
		{
			name: "大于条件",
			filters: []FilterCondition{
				{Field: "age", Operator: ">", Value: 18},
			},
			expected: 1,
		},
		{
			name: "LIKE条件",
			filters: []FilterCondition{
				{Field: "name", Operator: "LIKE", Value: "%test%"},
			},
			expected: 1,
		},
		{
			name: "多个条件",
			filters: []FilterCondition{
				{Field: "status", Operator: "=", Value: "active"},
				{Field: "amount", Operator: ">=", Value: 100},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditions := service.buildWhereConditions(tt.filters)
			assert.Equal(t, tt.expected, len(conditions))
		})
	}
}
