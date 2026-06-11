package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/engine"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/testutil"
	"cozy-insight/pkg/cache"
)

type mockChartConnector struct {
	queryCalled bool
	querySQL    string
	queryArgs   []interface{}
	returnData  []map[string]interface{}
}

func (m *mockChartConnector) Connect(string) error { return nil }
func (m *mockChartConnector) Close() error         { return nil }
func (m *mockChartConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	m.queryCalled = true
	m.querySQL = query
	m.queryArgs = args
	if m.returnData != nil {
		return m.returnData, nil
	}
	return []map[string]interface{}{}, nil
}
func (m *mockChartConnector) GetColumns(context.Context, string, string) ([]engine.ColumnInfo, error) {
	return nil, nil
}

func TestChartService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil, nil)

	mock.ExpectExec("INSERT INTO charts").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateChartRequest{
		Title:     "Test Chart",
		Type:      "bar",
		DatasetID: 1,
		Config:    "{}",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "Test Chart", result.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_Create_EmptyConfigDefaultsToBraces(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil, nil)

	mock.ExpectExec("INSERT INTO charts").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateChartRequest{
		Title:     "No Config",
		Type:      "line",
		DatasetID: 1,
		// Config intentionally empty
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, "{}", result.Config)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_Create_InvalidJSON(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil, nil)

	_, err := svc.Create(context.Background(), &dto.CreateChartRequest{
		Title:     "Bad",
		Type:      "bar",
		DatasetID: 1,
		Config:    `not-json`,
	}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

func TestChartService_Create_DBError(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil, nil)

	mock.ExpectExec("INSERT INTO charts").
		WillReturnError(sql.ErrConnDone)

	_, err := svc.Create(context.Background(), &dto.CreateChartRequest{
		Title:     "X",
		Type:      "bar",
		DatasetID: 1,
		Config:    "{}",
	}, 1)
	assert.Error(t, err)
}

func TestChartService_GetByID(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil, nil)

	columns := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test Chart", "bar", 1, "{}", 1, 1, now, now, nil,
		))

	chart, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), chart.ID)
	assert.Equal(t, "Test Chart", chart.Title)
	assert.Equal(t, "bar", chart.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_List(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil, nil)

	columns := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "Chart1", "bar", 1, "{}", 1, 1, now, now, nil))

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Chart1", list[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_Update(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil, nil)

	columns := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old", "bar", 1, "{}", 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE charts SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	status := int8(0)
	datasetID := uint64(2)
	err := svc.Update(context.Background(), 1, &dto.UpdateChartRequest{
		Title:     "Updated",
		Type:      "line",
		DatasetID: &datasetID,
		Config:    `{"color":"red"}`,
		Status:    &status,
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_Update_NotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil, nil)

	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	err := svc.Update(context.Background(), 1, &dto.UpdateChartRequest{Title: "test"})
	assert.Error(t, err)
}

func TestChartService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil, nil)

	mock.ExpectExec("UPDATE charts SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil, nil)

	// Mock chart SELECT
	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	// Mock dataset SELECT
	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(
			1, "Sales DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil,
		))

	// Mock datasource SELECT
	datasourceCols := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceCols).AddRow(
			1, "Local MySQL", "mysql", `{"host":"localhost","port":3306}`, 1, 1, now, now, nil,
		))

	// Connection will fail because config is incomplete (no username/password/database)
	_, err := svc.GetData(context.Background(), 1, nil, nil)
	assert.Error(t, err)
	// Should get "connect failed" error from Ping failing
}

func TestChartService_GetData_WithRuntimeFilters(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	pool := engine.NewConnectorPool()
	mockConn := &mockChartConnector{}
	pool.Register(1, mockConn)
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil, pool)

	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(
			1, "Sales DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil,
		))

	datasourceCols := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceCols).AddRow(
			1, "Local MySQL", "mysql", `{"host":"h","port":3306,"username":"u","password":"p","database":"db"}`, 1, 1, now, now, nil,
		))

	runtimeFilters := []dto.ChartFilter{
		{Field: "region", Operator: "=", Value: "US"},
	}

	_, err := svc.GetData(context.Background(), 1, runtimeFilters, nil)
	require.NoError(t, err)
	assert.True(t, mockConn.queryCalled)
	assert.Contains(t, mockConn.querySQL, "`region` = ?")
	assert.Contains(t, mockConn.queryArgs, "US")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData_WithDrillDimension(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	pool := engine.NewConnectorPool()
	mockConn := &mockChartConnector{}
	pool.Register(1, mockConn)
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil, pool)

	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(
			1, "Sales DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil,
		))

	datasourceCols := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceCols).AddRow(
			1, "Local MySQL", "mysql", `{"host":"h","port":3306,"username":"u","password":"p","database":"db"}`, 1, 1, now, now, nil,
		))

	drillDim := "region"
	_, err := svc.GetData(context.Background(), 1, nil, &drillDim)
	require.NoError(t, err)
	assert.True(t, mockConn.queryCalled)
	assert.Contains(t, mockConn.querySQL, "`region`")
	assert.NotContains(t, mockConn.querySQL, "`month`")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData_RuntimeParamsBypassCache(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	pool := engine.NewConnectorPool()
	mockConn := &mockChartConnector{
		returnData: []map[string]interface{}{{"region": "US", "sum_amount": int64(42)}},
	}
	pool.Register(1, mockConn)

	mr := miniredis.RunT(t)
	redisClient := cache.NewRedisClient(mr.Addr())
	cacheSvc := NewCacheService(redisClient)

	svc := NewChartService(chartRepo, datasetRepo, dsRepo, cacheSvc, pool)

	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(
			1, "Sales DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil,
		))

	datasourceCols := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceCols).AddRow(
			1, "Local MySQL", "mysql", `{"host":"h","port":3306,"username":"u","password":"p","database":"db"}`, 1, 1, now, now, nil,
		))

	chartConfig := `{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`
	cachedResp := &dto.ChartDataResponse{
		Dimensions: []string{"month"},
		Metrics:    []string{"sum_amount"},
		Data:       []map[string]interface{}{{`month`: "Jan", "sum_amount": int64(999)}},
	}
	err := cacheSvc.SetChartData(context.Background(), 1, chartConfig, cachedResp, time.Minute)
	require.NoError(t, err)

	runtimeFilters := []dto.ChartFilter{
		{Field: "region", Operator: "=", Value: "US"},
	}

	resp, err := svc.GetData(context.Background(), 1, runtimeFilters, nil)
	require.NoError(t, err)
	assert.True(t, mockConn.queryCalled, "connector should be called when cache is bypassed")
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "US", resp.Data[0]["region"])
	assert.Equal(t, int64(42), resp.Data[0]["sum_amount"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData_CacheHit(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	pool := engine.NewConnectorPool()
	mockConn := &mockChartConnector{}
	pool.Register(1, mockConn)

	mr := miniredis.RunT(t)
	redisClient := cache.NewRedisClient(mr.Addr())
	cacheSvc := NewCacheService(redisClient)

	svc := NewChartService(chartRepo, datasetRepo, dsRepo, cacheSvc, pool)

	chartConfig := `{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`
	cachedResp := &dto.ChartDataResponse{
		Dimensions: []string{"month"},
		Metrics:    []string{"sum_amount"},
		Data:       []map[string]interface{}{{`month`: "Jan", "sum_amount": int64(999)}},
	}
	err := cacheSvc.SetChartData(context.Background(), 1, chartConfig, cachedResp, time.Minute)
	require.NoError(t, err)

	// On cache hit the chart row is loaded for the cache key, but dataset / datasource /
	// connector must NOT be touched.
	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1, chartConfig, 1, 1, now, now, nil,
		))

	resp, err := svc.GetData(context.Background(), 1, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, float64(999), resp.Data[0]["sum_amount"], "JSON round-trip deserializes ints as float64")
	assert.False(t, mockConn.queryCalled, "connector must NOT be called on cache hit")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData_ChartNotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil, nil)

	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetData(context.Background(), 99, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chart not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData_InvalidConfigJSON(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil, nil)

	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Bad", "bar", 1, "{not-json", 1, 1, now, now, nil,
		))

	_, err := svc.GetData(context.Background(), 1, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid chart config")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData_DatasetNotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	pool := engine.NewConnectorPool()
	pool.Register(1, &mockChartConnector{})
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil, pool)

	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetData(context.Background(), 1, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dataset not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData_DatasourceNotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	pool := engine.NewConnectorPool()
	pool.Register(1, &mockChartConnector{})
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil, pool)

	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(
			1, "Sales DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil,
		))

	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetData(context.Background(), 1, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "datasource not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData_QueryError(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	pool := engine.NewConnectorPool()
	failingConn := &failingQueryConnector{err: fmt.Errorf("kaboom")}
	pool.Register(1, failingConn)
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil, pool)

	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(
			1, "Sales DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil,
		))

	datasourceCols := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceCols).AddRow(
			1, "Local MySQL", "mysql", `{"host":"h","port":3306,"username":"u","password":"p","database":"db"}`, 1, 1, now, now, nil,
		))

	_, err := svc.GetData(context.Background(), 1, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartService_GetData_DrillDimension_OverwritesConfig(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	pool := engine.NewConnectorPool()
	mockConn := &mockChartConnector{
		returnData: []map[string]interface{}{{"day": "Mon", "sum_amount": int64(1)}},
	}
	pool.Register(1, mockConn)
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil, pool)

	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(
			1, "Sales DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil,
		))

	datasourceCols := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceCols).AddRow(
			1, "Local MySQL", "mysql", `{"host":"h","port":3306,"username":"u","password":"p","database":"db"}`, 1, 1, now, now, nil,
		))

	drill := "day"
	resp, err := svc.GetData(context.Background(), 1, nil, &drill)
	require.NoError(t, err)
	assert.Equal(t, []string{"day"}, resp.Dimensions, "drill dimension must replace configured dimensions")
	assert.Contains(t, mockConn.querySQL, "`day`")
	assert.NotContains(t, mockConn.querySQL, "`month`")
	assert.NoError(t, mock.ExpectationsWereMet())
}

type failingQueryConnector struct {
	err error
}

func (m *failingQueryConnector) Connect(string) error { return nil }
func (m *failingQueryConnector) Close() error         { return nil }
func (m *failingQueryConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return nil, m.err
}
func (m *failingQueryConnector) GetColumns(context.Context, string, string) ([]engine.ColumnInfo, error) {
	return nil, nil
}
