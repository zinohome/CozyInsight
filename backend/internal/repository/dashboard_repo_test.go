package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/model"
)

func TestDashboardRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO dashboards").
		WillReturnResult(sqlmock.NewResult(7, 1))

	d := &model.Dashboard{Title: "Main", Type: "dashboard", Config: "{}", Status: 1, CreatedBy: 1}
	require.NoError(t, repo.Create(context.Background(), d))
	assert.Equal(t, uint64(7), d.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "Main", "dashboard", "{}", "", 0, 1, 1, now, now))

	d, err := repo.FindByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Main", d.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "Main", "dashboard", "{}", "", 0, 1, 1, now, now))

	ds, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, ds, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE dashboards SET title").
		WillReturnResult(sqlmock.NewResult(0, 1))

	d := &model.Dashboard{ID: 1, Title: "New", Type: "dashboard", Config: "{}", Status: 1}
	require.NoError(t, repo.Update(context.Background(), d))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_FindByShareToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE share_token = \\? AND deleted_at IS NULL").
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "Main", "dashboard", "{}", "abc123", 1, 1, 1, now, now))

	d, err := repo.FindByShareToken(context.Background(), "abc123")
	require.NoError(t, err)
	assert.Equal(t, "abc123", d.ShareToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE dashboards SET deleted_at").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_AddChart(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO dashboard_charts").
		WillReturnResult(sqlmock.NewResult(1, 1))

	dc := &model.DashboardChart{DashboardID: 1, ChartID: 2, PositionX: 0, PositionY: 0, Width: 6, Height: 4, Config: "{}"}
	require.NoError(t, repo.AddChart(context.Background(), dc))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetCharts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "dashboard_id", "chart_id", "position_x", "position_y", "width", "height", "config"}
	mock.ExpectQuery("SELECT \\* FROM dashboard_charts WHERE dashboard_id = \\? ORDER BY id").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 1, 2, 0, 0, 6, 4, "{}"))

	dcs, err := repo.GetCharts(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, dcs, 1)
	assert.Equal(t, uint64(2), dcs[0].ChartID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_RemoveChart(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("DELETE FROM dashboard_charts WHERE dashboard_id = \\? AND chart_id = \\?").
		WithArgs(uint64(1), uint64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.RemoveChart(context.Background(), 1, 2))
	assert.NoError(t, mock.ExpectationsWereMet())
}
