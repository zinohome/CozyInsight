package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/testutil"
)

func TestDashboardService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo)

	mock.ExpectExec("INSERT INTO dashboards").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateDashboardRequest{
		Title:  "Test Dashboard",
		Config: "{}",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "Test Dashboard", result.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_GetByID(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo)

	columns := []string{"id", "title", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test Dashboard", "{}", 1, 1, now, now, nil,
		))

	d, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), d.ID)
	assert.Equal(t, "Test Dashboard", d.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_List(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo)

	columns := []string{"id", "title", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "Dashboard1", "{}", 1, 1, now, now, nil).
			AddRow(2, "Dashboard2", "{}", 1, 1, now, now, nil))

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Update(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo)

	columns := []string{"id", "title", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old", "{}", 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE dashboards SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	status := int8(0)
	err := svc.Update(context.Background(), 1, &dto.UpdateDashboardRequest{
		Title:  "Updated",
		Config: `{"color":"blue"}`,
		Status: &status,
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo)

	mock.ExpectExec("UPDATE dashboards SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_AddChart(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo)

	mock.ExpectExec("INSERT INTO dashboard_charts").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.AddChart(context.Background(), 1, &dto.AddChartToDashboardRequest{
		ChartID:   1,
		PositionX: 0,
		PositionY: 0,
		Width:     6,
		Height:    4,
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_GetCharts(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo)

	columns := []string{"id", "dashboard_id", "chart_id", "position_x", "position_y", "width", "height", "config", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboard_charts WHERE dashboard_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, 1, 1, 0, 0, 6, 4, "{}", now, now,
		))

	list, err := svc.GetCharts(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, uint64(1), list[0].ChartID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_RemoveChart(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo)

	mock.ExpectExec("DELETE FROM dashboard_charts WHERE dashboard_id = \\? AND chart_id = \\?").
		WithArgs(1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.RemoveChart(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
