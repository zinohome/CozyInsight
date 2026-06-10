package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkbenchRepository_CountByCreatedBy(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkbenchRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM dashboards WHERE created_by = \\? AND deleted_at IS NULL").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	c, err := repo.CountByCreatedBy(context.Background(), "dashboards", 1, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), c)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchRepository_CountByCreatedBy_WithExtra(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkbenchRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM charts WHERE created_by = \\? AND deleted_at IS NULL AND status = 1").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	c, err := repo.CountByCreatedBy(context.Background(), "charts", 1, "status = 1")
	require.NoError(t, err)
	assert.Equal(t, int64(0), c)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchRepository_UpsertRecentView(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkbenchRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO recent_views .* ON DUPLICATE KEY UPDATE visited_at").
		WithArgs(uint64(1), "dashboard", uint64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UpsertRecentView(context.Background(), 1, "dashboard", 5))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchRepository_ListRecentViews(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkbenchRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "title", "type", "visited_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT d.id, d.title, d.type, rv.visited_at .* LIMIT \\?").
		WithArgs(uint64(1), 10).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "Dash", "dashboard", now))

	views, err := repo.ListRecentViews(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Len(t, views, 1)
	assert.Equal(t, "Dash", views[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchRepository_AddFavorite(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkbenchRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO favorites .* ON DUPLICATE KEY UPDATE created_at").
		WithArgs(uint64(1), "dashboard", uint64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.AddFavorite(context.Background(), 1, "dashboard", 2))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchRepository_DeleteFavorite(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkbenchRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("DELETE FROM favorites WHERE user_id = \\? AND resource_type = \\? AND resource_id = \\?").
		WithArgs(uint64(1), "dashboard", uint64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.DeleteFavorite(context.Background(), 1, "dashboard", 2))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkbenchRepository_ListFavorites(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewWorkbenchRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "title", "type", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT d.id, d.title, d.type, f.created_at .*").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "Dash", "dashboard", now))

	favs, err := repo.ListFavorites(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, favs, 1)
	assert.Equal(t, "Dash", favs[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}
