package service

import (
	"context"
	"cozy-insight-backend/internal/engine"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatasetRepo 用于测试的Mock Repository
type MockDatasetRepo struct {
	mock.Mock
}

func (m *MockDatasetRepo) CreateGroup(ctx context.Context, group *model.DatasetGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockDatasetRepo) ListGroups(ctx context.Context) ([]*model.DatasetGroup, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.DatasetGroup), args.Error(1)
}

func (m *MockDatasetRepo) CreateTable(ctx context.Context, table *model.DatasetTable) error {
	args := m.Called(ctx, table)
	return args.Error(0)
}

func (m *MockDatasetRepo) UpdateTable(ctx context.Context, table *model.DatasetTable) error {
	args := m.Called(ctx, table)
	return args.Error(0)
}

func (m *MockDatasetRepo) GetTable(ctx context.Context, id string) (*model.DatasetTable, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DatasetTable), args.Error(1)
}

func (m *MockDatasetRepo) ListTables(ctx context.Context, groupId string) ([]*model.DatasetTable, error) {
	args := m.Called(ctx, groupId)
	return args.Get(0).([]*model.DatasetTable), args.Error(1)
}

func (m *MockDatasetRepo) SaveFields(ctx context.Context, fields []*model.DatasetTableField) error {
	args := m.Called(ctx, fields)
	return args.Error(0)
}

func (m *MockDatasetRepo) GetFields(ctx context.Context, tableId string) ([]*model.DatasetTableField, error) {
	args := m.Called(ctx, tableId)
	return args.Get(0).([]*model.DatasetTableField), args.Error(1)
}

func (m *MockDatasetRepo) DeleteFieldsByTableID(ctx context.Context, tableId string) error {
	args := m.Called(ctx, tableId)
	return args.Error(0)
}

func (m *MockDatasetRepo) BatchCreateFields(ctx context.Context, fields []*model.DatasetTableField) error {
	args := m.Called(ctx, fields)
	return args.Error(0)
}

// MockCalciteClient 用于测试的Mock Calcite Client
type MockCalciteClient struct {
	mock.Mock
}

func (m *MockCalciteClient) ExecuteQuery(ctx context.Context, sql string, params ...interface{}) ([]map[string]interface{}, error) {
	args := m.Called(ctx, sql, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockCalciteClient) ExecuteQueryNoCache(ctx context.Context, sql string, params ...interface{}) ([]map[string]interface{}, error) {
	args := m.Called(ctx, sql, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// Test CreateGroup
func TestDatasetService_CreateGroup(t *testing.T) {
	mockRepo := new(MockDatasetRepo)
	service := NewDatasetService(mockRepo, nil)

	group := &model.DatasetGroup{
		ID:   "group1",
		Name: "Test Group",
		PID:  "root",
	}

	mockRepo.On("CreateGroup", mock.Anything, group).Return(nil)

	err := service.CreateGroup(context.Background(), group)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test ListGroups
func TestDatasetService_ListGroups(t *testing.T) {
	mockRepo := new(MockDatasetRepo)
	service := NewDatasetService(mockRepo, nil)

	expectedGroups := []*model.DatasetGroup{
		{ID: "group1", Name: "Group 1"},
		{ID: "group2", Name: "Group 2"},
	}

	mockRepo.On("ListGroups", mock.Anything).Return(expectedGroups, nil)

	groups, err := service.ListGroups(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(groups))
	assert.Equal(t, "Group 1", groups[0].Name)
	mockRepo.AssertExpectations(t)
}

// Test CreateTable
func TestDatasetService_CreateTable(t *testing.T) {
	mockRepo := new(MockDatasetRepo)
	service := NewDatasetService(mockRepo, nil)

	table := &model.DatasetTable{
		ID:                "table1",
		Name:              "Test Table",
		PhysicalTableName: "test_table",
		Type:              "db",
	}

	mockRepo.On("CreateTable", mock.Anything, table).Return(nil)

	err := service.CreateTable(context.Background(), table)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test PreviewData
func TestDatasetService_PreviewData(t *testing.T) {
	mockRepo := new(MockDatasetRepo)
	mockCalcite := new(MockCalciteClient)
	service := NewDatasetService(mockRepo, mockCalcite)

	table := &model.DatasetTable{
		ID:                "table1",
		Name:              "Test Table",
		PhysicalTableName: "sales",
		Type:              "db",
		DatasourceID:      "ds1",
	}

	// Mock数据
	mockData := []map[string]interface{}{
		{
			"id":       1,
			"name":     "Product A",
			"price":    99.99,
			"quantity": 10,
			"date":     "2024-01-01",
		},
		{
			"id":       2,
			"name":     "Product B",
			"price":    149.99,
			"quantity": 5,
			"date":     "2024-01-02",
		},
	}

	mockRepo.On("GetTable", mock.Anything, "table1").Return(table, nil)
	mockCalcite.On("ExecuteQuery", mock.Anything, mock.MatchedBy(func(sql string) bool {
		return true // 匹配任何SQL
	}), mock.Anything).Return(mockData, nil)

	result, err := service.PreviewData(context.Background(), "table1", 10)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 5, len(result.Fields)) // 应该推断出5个字段
	assert.Equal(t, 2, len(result.Data))

	mockRepo.AssertExpectations(t)
	mockCalcite.AssertExpectations(t)
}

// Test InferFieldTypes - 字段类型推断
func TestDatasetService_InferFieldTypes(t *testing.T) {
	service := &datasetService{}

	testData := []map[string]interface{}{
		{
			"id":        123,
			"name":      "Test",
			"price":     99.99,
			"active":    true,
			"date":      "2024-01-01",
			"timestamp": time.Now(),
		},
	}

	fields := service.inferFieldTypes(testData)

	assert.Equal(t, 6, len(fields))

	// 验证字段类型推断
	fieldMap := make(map[string]*FieldInfo)
	for _, field := range fields {
		fieldMap[field.Name] = field
	}

	// 整数字段
	assert.Equal(t, "INT", fieldMap["id"].Type)
	assert.Equal(t, 2, fieldMap["id"].DeType) // int类型
	assert.Equal(t, "q", fieldMap["id"].GroupType) // measure

	// 字符串字段
	assert.Equal(t, "VARCHAR", fieldMap["name"].Type)
	assert.Equal(t, 0, fieldMap["name"].DeType) // text类型
	assert.Equal(t, "d", fieldMap["name"].GroupType) // dimension

	// 浮点数字段
	assert.Equal(t, "DECIMAL", fieldMap["price"].Type)
	assert.Equal(t, 3, fieldMap["price"].DeType) // float类型

	// 布尔字段
	assert.Equal(t, "BOOLEAN", fieldMap["active"].Type)
	assert.Equal(t, 4, fieldMap["active"].DeType) // bool类型
}

// Test DetectFieldType - 详细类型检测
func TestDatasetService_DetectFieldType(t *testing.T) {
	service := &datasetService{}

	tests := []struct {
		name          string
		value         interface{}
		expectedType  string
		expectedDeType int
		expectedGroup string
	}{
		{"nil value", nil, "VARCHAR", 0, "d"},
		{"integer", 123, "INT", 2, "q"},
		{"int64", int64(456), "BIGINT", 2, "q"},
		{"float32", float32(99.9), "DECIMAL", 3, "q"},
		{"float64", 99.99, "DECIMAL", 3, "q"},
		{"boolean true", true, "BOOLEAN", 4, "d"},
		{"boolean false", false, "BOOLEAN", 4, "d"},
		{"string", "test", "VARCHAR", 0, "d"},
		{"date string", "2024-01-01", "DATE", 1, "d"},
		{"datetime string", "2024-01-01 12:00:00", "DATETIME", 1, "d"},
		{"time.Time", time.Now(), "DATETIME", 1, "d"},
		{"bytes", []byte("data"), "VARCHAR", 0, "d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbType, deType, groupType := service.detectFieldType(tt.value)
			assert.Equal(t, tt.expectedType, dbType)
			assert.Equal(t, tt.expectedDeType, deType)
			assert.Equal(t, tt.expectedGroup, groupType)
		})
	}
}

// Test SyncFields
func TestDatasetService_SyncFields(t *testing.T) {
	mockRepo := new(MockDatasetRepo)
	mockCalcite := new(MockCalciteClient)
	service := NewDatasetService(mockRepo, mockCalcite)

	table := &model.DatasetTable{
		ID:                "table1",
		Name:              "Test Table",
		PhysicalTableName: "sales",
		Type:              "db",
		DatasourceID:      "ds1",
		DatasetGroupID:    "group1",
	}

	mockData := []map[string]interface{}{
		{"id": 1, "name": "Product A", "price": 99.99},
	}

	mockRepo.On("GetTable", mock.Anything, "table1").Return(table, nil)
	mockCalcite.On("ExecuteQuery", mock.Anything, mock.Anything, mock.Anything).Return(mockData, nil)
	mockRepo.On("DeleteFieldsByTableID", mock.Anything, "table1").Return(nil)
	mockRepo.On("BatchCreateFields", mock.Anything, mock.MatchedBy(func(fields []*model.DatasetTableField) bool {
		return len(fields) == 3 // 应该有3个字段
	})).Return(nil)

	err := service.SyncFields(context.Background(), "table1")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCalcite.AssertExpectations(t)
}

// Test BuildPreviewSQL
func TestDatasetService_BuildPreviewSQL(t *testing.T) {
	service := &datasetService{}

	tests := []struct {
		name      string
		table     *model.DatasetTable
		limit     int
		expectSQL string
	}{
		{
			name: "db type table",
			table: &model.DatasetTable{
				Type:              "db",
				PhysicalTableName: "orders",
			},
			limit:     10,
			expectSQL: "SELECT * FROM orders LIMIT 10",
		},
		{
			name: "sql type table",
			table: &model.DatasetTable{
				Type: "sql",
				Info: `{"sql":"SELECT * FROM custom_view"}`,
			},
			limit:     20,
			expectSQL: "SELECT * FROM (SELECT * FROM custom_view) AS preview LIMIT 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := service.buildPreviewSQL(tt.table, tt.limit)
			assert.NoError(t, err)
			assert.Contains(t, sql, "SELECT")
			assert.Contains(t, sql, "LIMIT")
		})
	}
}
