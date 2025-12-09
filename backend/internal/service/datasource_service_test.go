package service

import (
	"context"
	"cozy-insight-backend/internal/engine"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatasourceRepo for testing
type MockDatasourceRepo struct {
	mock.Mock
}

func (m *MockDatasourceRepo) Create(ctx context.Context, ds *model.Datasource) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *MockDatasourceRepo) Update(ctx context.Context, ds *model.Datasource) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *MockDatasourceRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDatasourceRepo) GetByID(ctx context.Context, id string) (*model.Datasource, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Datasource), args.Error(1)
}

func (m *MockDatasourceRepo) List(ctx context.Context) ([]*model.Datasource, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Datasource), args.Error(1)
}

// MockDatasourceConnector for testing
type MockDatasourceConnector struct {
	mock.Mock
}

func (m *MockDatasourceConnector) TestConnection(ctx context.Context, configJSON string) (*engine.ConnectionTestResult, error) {
	args := m.Called(ctx, configJSON)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*engine.ConnectionTestResult), args.Error(1)
}

func (m *MockDatasourceConnector) GetDatabaseList(ctx context.Context, configJSON string) ([]string, error) {
	args := m.Called(ctx, configJSON)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDatasourceConnector) GetTableList(ctx context.Context, configJSON, database string) ([]string, error) {
	args := m.Called(ctx, configJSON, database)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDatasourceConnector) GetTableSchema(ctx context.Context, configJSON, database, table string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, configJSON, database, table)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// Test Create
func TestDatasourceService_Create(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	ds := &model.Datasource{
		Name:          "MySQL Test",
		Type:          "mysql",
		Configuration: `{"host":"localhost","port":3306}`,
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(d *model.Datasource) bool {
		return d.ID != "" && d.CreateTime > 0
	})).Return(nil)

	err := service.Create(context.Background(), ds)
	assert.NoError(t, err)
	assert.NotEmpty(t, ds.ID)
	assert.NotZero(t, ds.CreateTime)
	mockRepo.AssertExpectations(t)
}

// Test Update
func TestDatasourceService_Update(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	ds := &model.Datasource{
		ID:            "ds123",
		Name:          "Updated MySQL",
		Type:          "mysql",
		Configuration: `{"host":"newhost","port":3306}`,
		CreateTime:    1000,
	}

	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(d *model.Datasource) bool {
		return d.UpdateTime > d.CreateTime
	})).Return(nil)

	err := service.Update(context.Background(), ds)
	assert.NoError(t, err)
	assert.Greater(t, ds.UpdateTime, ds.CreateTime)
	mockRepo.AssertExpectations(t)
}

// Test Delete
func TestDatasourceService_Delete(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	mockRepo.On("Delete", mock.Anything, "ds123").Return(nil)

	err := service.Delete(context.Background(), "ds123")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test GetByID
func TestDatasourceService_GetByID(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	expectedDS := &model.Datasource{
		ID:   "ds123",
		Name: "Test MySQL",
		Type: "mysql",
	}

	mockRepo.On("GetByID", mock.Anything, "ds123").Return(expectedDS, nil)

	ds, err := service.GetByID(context.Background(), "ds123")
	assert.NoError(t, err)
	assert.Equal(t, "Test MySQL", ds.Name)
	mockRepo.AssertExpectations(t)
}

// Test List
func TestDatasourceService_List(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	expectedList := []*model.Datasource{
		{ID: "ds1", Name: "MySQL", Type: "mysql"},
		{ID: "ds2", Name: "PostgreSQL", Type: "postgresql"},
	}

	mockRepo.On("List", mock.Anything).Return(expectedList, nil)

	list, err := service.List(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(list))
	mockRepo.AssertExpectations(t)
}

// Test TestConnection
func TestDatasourceService_TestConnection(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	ds := &model.Datasource{
		ID:            "ds123",
		Configuration: `{"host":"localhost","port":3306}`,
	}

	expectedResult := &engine.ConnectionTestResult{
		Success: true,
		Message: "Connection successful",
		Latency: 50,
	}

	mockRepo.On("GetByID", mock.Anything, "ds123").Return(ds, nil)
	mockConnector.On("TestConnection", mock.Anything, ds.Configuration).Return(expectedResult, nil)

	result, err := service.TestConnection(context.Background(), "ds123")
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "Connection successful", result.Message)
	mockRepo.AssertExpectations(t)
	mockConnector.AssertExpectations(t)
}

// Test TestConnectionByConfig
func TestDatasourceService_TestConnectionByConfig(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	config := `{"host":"localhost","port":3306,"username":"root"}`

	expectedResult := &engine.ConnectionTestResult{
		Success: true,
		Message: "Connection successful",
		Latency: 30,
	}

	mockConnector.On("TestConnection", mock.Anything, config).Return(expectedResult, nil)

	result, err := service.TestConnectionByConfig(context.Background(), config)
	assert.NoError(t, err)
	assert.True(t, result.Success)
	mockConnector.AssertExpectations(t)
}

// Test GetDatabases
func TestDatasourceService_GetDatabases(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	ds := &model.Datasource{
		ID:            "ds123",
		Configuration: `{"host":"localhost"}`,
	}

	expectedDatabases := []string{"db1", "db2", "db3"}

	mockRepo.On("GetByID", mock.Anything, "ds123").Return(ds, nil)
	mockConnector.On("GetDatabaseList", mock.Anything, ds.Configuration).Return(expectedDatabases, nil)

	databases, err := service.GetDatabases(context.Background(), "ds123")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(databases))
	assert.Contains(t, databases, "db1")
	mockRepo.AssertExpectations(t)
	mockConnector.AssertExpectations(t)
}

// Test GetTables
func TestDatasourceService_GetTables(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	ds := &model.Datasource{
		ID:            "ds123",
		Configuration: `{"host":"localhost"}`,
	}

	expectedTables := []string{"table1", "table2", "users", "orders"}

	mockRepo.On("GetByID", mock.Anything, "ds123").Return(ds, nil)
	mockConnector.On("GetTableList", mock.Anything, ds.Configuration, "testdb").Return(expectedTables, nil)

	tables, err := service.GetTables(context.Background(), "ds123", "testdb")
	assert.NoError(t, err)
	assert.Equal(t, 4, len(tables))
	assert.Contains(t, tables, "users")
	mockRepo.AssertExpectations(t)
	mockConnector.AssertExpectations(t)
}

// Test GetTableSchema
func TestDatasourceService_GetTableSchema(t *testing.T) {
	mockRepo := new(MockDatasourceRepo)
	mockConnector := new(MockDatasourceConnector)
	service := NewDatasourceService(mockRepo, mockConnector)

	ds := &model.Datasource{
		ID:            "ds123",
		Configuration: `{"host":"localhost"}`,
	}

	expectedSchema := []map[string]interface{}{
		{"column_name": "id", "data_type": "int", "is_nullable": "NO"},
		{"column_name": "name", "data_type": "varchar", "is_nullable": "YES"},
	}

	mockRepo.On("GetByID", mock.Anything, "ds123").Return(ds, nil)
	mockConnector.On("GetTableSchema", mock.Anything, ds.Configuration, "testdb", "users").Return(expectedSchema, nil)

	schema, err := service.GetTableSchema(context.Background(), "ds123", "testdb", "users")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(schema))
	assert.Equal(t, "id", schema[0]["column_name"])
	mockRepo.AssertExpectations(t)
	mockConnector.AssertExpectations(t)
}
