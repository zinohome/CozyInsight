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

func TestDatasetRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasetRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO datasets").
		WillReturnResult(sqlmock.NewResult(5, 1))

	ds := &model.Dataset{Name: "ds1", DatasourceID: 1, DatabaseName: "db", TableName: "t", SQL: "", Type: "table", Mode: 1, Status: 1, CreatedBy: 1}
	require.NoError(t, repo.Create(context.Background(), ds))
	assert.Equal(t, uint64(5), ds.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasetRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "ds1", 1, "db", "t", "", "table", 1, 1, 1, now, now))

	ds, err := repo.FindByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "ds1", ds.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasetRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "sql", "type", "mode", "status", "created_by", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "ds1", 1, "db", "t", "", "table", 1, 1, 1, now, now))

	dss, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, dss, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasetRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE datasets SET name").
		WillReturnResult(sqlmock.NewResult(0, 1))

	ds := &model.Dataset{ID: 1, Name: "new", DatasourceID: 1, DatabaseName: "db", TableName: "t", SQL: "", Type: "table", Mode: 1, Status: 1}
	require.NoError(t, repo.Update(context.Background(), ds))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasetRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE datasets SET deleted_at").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetRepository_CreateFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasetRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO dataset_fields").
		WillReturnResult(sqlmock.NewResult(0, 2))

	fields := []model.DatasetField{
		{DatasetID: 1, Name: "id", Type: "int", DeType: 0, Length: 11},
		{DatasetID: 1, Name: "name", Type: "varchar", DeType: 0, Length: 64},
	}
	require.NoError(t, repo.CreateFields(context.Background(), fields))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetRepository_GetFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasetRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "dataset_id", "name", "type", "de_type", "length", "precision", "scale", "origin_name"}
	mock.ExpectQuery("SELECT \\* FROM dataset_fields WHERE dataset_id = \\? ORDER BY id").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 1, "id", "int", 0, 11, 0, 0, "id"))

	fields, err := repo.GetFields(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, fields, 1)
	assert.Equal(t, "id", fields[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetRepository_DeleteFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDatasetRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("DELETE FROM dataset_fields WHERE dataset_id = \\?").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.DeleteFields(context.Background(), 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}
