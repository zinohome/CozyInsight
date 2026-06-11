package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/repository"
	"cozy-insight/internal/testutil"
)

func TestWorkbenchService_GetStats(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewWorkbenchRepository(db)
	svc := NewWorkbenchService(repo)

	// Five sequential CountByCreatedBy calls.
	// 1: datasources  2: datasets  3: charts  4: dashboards  5: dashboards (screen)
	mock.ExpectQuery("FROM datasources WHERE created_by = \\? AND deleted_at IS NULL").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(3))
	mock.ExpectQuery("FROM datasets WHERE created_by = \\? AND deleted_at IS NULL").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(5))
	mock.ExpectQuery("FROM charts WHERE created_by = \\? AND deleted_at IS NULL").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(8))
	mock.ExpectQuery("FROM dashboards WHERE created_by = \\? AND deleted_at IS NULL AND type = 'dashboard'").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(2))
	mock.ExpectQuery("FROM dashboards WHERE created_by = \\? AND deleted_at IS NULL AND type = 'screen'").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(1))

	stats, err := svc.GetStats(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, int64(3), stats.DatasourceCount)
	assert.Equal(t, int64(5), stats.DatasetCount)
	assert.Equal(t, int64(8), stats.ChartCount)
	assert.Equal(t, int64(2), stats.DashboardCount)
	assert.Equal(t, int64(1), stats.ScreenCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchService_GetStats_DBError_Datasource(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewWorkbenchRepository(db)
	svc := NewWorkbenchService(repo)

	mock.ExpectQuery("FROM datasources").
		WithArgs(7).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.GetStats(context.Background(), 7)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "datasource count")
}

func TestWorkbenchService_RecordVisit(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewWorkbenchRepository(db)
	svc := NewWorkbenchService(repo)

	mock.ExpectExec("INSERT INTO recent_views").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.RecordVisit(context.Background(), 7, "dashboard", 100)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchService_ListRecentViews(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewWorkbenchRepository(db)
	svc := NewWorkbenchService(repo)

	now := time.Now()
	mock.ExpectQuery("FROM recent_views rv").
		WithArgs(7, 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "type", "visited_at"}).
			AddRow(1, "Sales Report", "dashboard", now).
			AddRow(2, "Marketing KPIs", "screen", now))

	list, err := svc.ListRecentViews(context.Background(), 7)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "Sales Report", list[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchService_AddFavorite(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewWorkbenchRepository(db)
	svc := NewWorkbenchService(repo)

	mock.ExpectExec("INSERT INTO favorites").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.AddFavorite(context.Background(), 7, "dashboard", 100)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchService_DeleteFavorite(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewWorkbenchRepository(db)
	svc := NewWorkbenchService(repo)

	mock.ExpectExec("DELETE FROM favorites").
		WithArgs(7, "dashboard", 100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.DeleteFavorite(context.Background(), 7, "dashboard", 100)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchService_ListFavorites(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewWorkbenchRepository(db)
	svc := NewWorkbenchService(repo)

	now := time.Now()
	mock.ExpectQuery("FROM favorites f").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "type", "created_at"}).
			AddRow(1, "Q1 Plan", "dashboard", now))

	list, err := svc.ListFavorites(context.Background(), 7)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Q1 Plan", list[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}
