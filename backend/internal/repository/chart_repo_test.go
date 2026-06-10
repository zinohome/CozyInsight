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

func TestChartRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewChartRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO charts").
		WillReturnResult(sqlmock.NewResult(42, 1))

	c := &model.Chart{Title: "Sales", Type: "bar", DatasetID: 1, Config: "{}", Status: 1, CreatedBy: 1}
	require.NoError(t, repo.Create(context.Background(), c))
	assert.Equal(t, uint64(42), c.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewChartRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "Sales", "bar", 1, "{}", 1, 1, now, now))

	c, err := repo.FindByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Sales", c.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewChartRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "a", "bar", 1, "{}", 1, 1, now, now))

	cs, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewChartRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE charts SET title").
		WillReturnResult(sqlmock.NewResult(0, 1))

	c := &model.Chart{ID: 1, Title: "New", Type: "line", DatasetID: 1, Config: "{}", Status: 1}
	require.NoError(t, repo.Update(context.Background(), c))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChartRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewChartRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE charts SET deleted_at").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}
