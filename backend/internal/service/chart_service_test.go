package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/testutil"
)

func TestChartService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil)

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

func TestChartService_GetByID(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(repo, datasetRepo, dsRepo, nil)

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
	svc := NewChartService(repo, datasetRepo, dsRepo, nil)

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
	svc := NewChartService(repo, datasetRepo, dsRepo, nil)

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
	svc := NewChartService(repo, datasetRepo, dsRepo, nil)

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
	svc := NewChartService(repo, datasetRepo, dsRepo, nil)

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
	svc := NewChartService(chartRepo, datasetRepo, dsRepo, nil)

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
	_, err := svc.GetData(context.Background(), 1)
	assert.Error(t, err)
	// Should get "connect failed" error from Ping failing
}
