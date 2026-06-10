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

func TestDatasourceRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasourceRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO datasources").
		WillReturnResult(sqlmock.NewResult(3, 1))

	ds := &model.Datasource{Name: "main", Type: "mysql", Config: `{"host":"x"}`, Status: 1, CreatedBy: 1}
	require.NoError(t, repo.Create(context.Background(), ds))
	assert.Equal(t, uint64(3), ds.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasourceRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "name", "type", "config", "file_path", "file_type", "status", "created_by", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "main", "mysql", "{}", "", "", 1, 1, now, now))

	ds, err := repo.FindByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "main", ds.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasourceRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "name", "type", "config", "file_path", "file_type", "status", "created_by", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "main", "mysql", "{}", "", "", 1, 1, now, now))

	dss, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, dss, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasourceRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE datasources SET name").
		WillReturnResult(sqlmock.NewResult(0, 1))

	ds := &model.Datasource{ID: 1, Name: "renamed", Type: "mysql", Config: "{}", Status: 1}
	require.NoError(t, repo.Update(context.Background(), ds))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasourceRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE datasources SET deleted_at").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}
