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

func TestDatasourceService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	mock.ExpectExec("INSERT INTO datasources").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateDatasourceRequest{
		Name:   "Test Datasource",
		Type:   "mysql",
		Config: `{"host":"localhost"}`,
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "Test Datasource", result.Name)
	assert.Equal(t, "mysql", result.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceService_GetByID(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	columns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test", "mysql", "{}", 1, 1, now, now, nil,
		))

	ds, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), ds.ID)
	assert.Equal(t, "Test", ds.Name)
	assert.Equal(t, "mysql", ds.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceService_GetByID_NotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByID(context.Background(), 999)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceService_List(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	columns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "DS1", "mysql", "{}", 1, 1, now, now, nil).
			AddRow(2, "DS2", "postgresql", "{}", 1, 1, now, now, nil))

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "DS1", list[0].Name)
	assert.Equal(t, "DS2", list[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceService_Update(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	columns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old", "mysql", "{}", 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE datasources SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	status := int8(0)
	err := svc.Update(context.Background(), 1, &dto.UpdateDatasourceRequest{
		Name:   "Updated",
		Type:   "postgresql",
		Config: `{"host":"new"}`,
		Status: &status,
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	columns := []string{"id", "name", "type", "config", "file_path", "file_type", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "MySQL", "mysql", "{}", "", "", 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE datasources SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceService_TestConnection_UnsupportedType(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	err := svc.TestConnection(context.Background(), &dto.TestConnectionRequest{
		Type: "unknown",
		Config: map[string]interface{}{
			"host": "localhost",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestDatasourceService_Create_InvalidJSON(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	_, err := svc.Create(context.Background(), &dto.CreateDatasourceRequest{
		Name:   "Bad Datasource",
		Type:   "mysql",
		Config: `not-json`,
	}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

func TestDatasourceService_Create_FileType(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	mock.ExpectExec("INSERT INTO datasources").
		WillReturnResult(sqlmock.NewResult(2, 1))

	result, err := svc.Create(context.Background(), &dto.CreateDatasourceRequest{
		Name:   "Sales CSV",
		Type:   "csv",
		Config: `{"file_path":"/data/sales.csv","file_type":"csv"}`,
	}, 7)
	require.NoError(t, err)
	assert.Equal(t, "Sales CSV", result.Name)
	assert.Equal(t, "csv", result.Type)
	assert.Equal(t, "/data/sales.csv", result.FilePath)
	assert.Equal(t, "csv", result.FileType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceService_Create_DBError(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	mock.ExpectExec("INSERT INTO datasources").
		WillReturnError(sql.ErrConnDone)

	_, err := svc.Create(context.Background(), &dto.CreateDatasourceRequest{
		Name:   "X",
		Type:   "mysql",
		Config: `{}`,
	}, 1)
	assert.Error(t, err)
}

func TestDatasourceService_TestConnection_MySQL_PingFails(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	// 127.0.0.1:1 is reserved (tcpmux) and universally refuses connections, so
	// sql.Open succeeds but PingContext returns an error.
	err := svc.TestConnection(context.Background(), &dto.TestConnectionRequest{
		Type: "mysql",
		Config: map[string]interface{}{
			"host":     "127.0.0.1",
			"port":     float64(1),
			"username": "u",
			"password": "p",
			"database": "d",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ping failed")
}

func TestDatasourceService_TestConnection_Postgres_OpenFails(t *testing.T) {
	// postgresql driver is registered as "postgres" in database/sql; the
	// service's "postgresql" branch passes req.Type through unchanged, so
	// sql.Open must reject the unknown driver.
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	err := svc.TestConnection(context.Background(), &dto.TestConnectionRequest{
		Type: "postgresql",
		Config: map[string]interface{}{
			"host":     "127.0.0.1",
			"port":     float64(1),
			"username": "u",
			"password": "p",
			"database": "d",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open connection failed")
}

func TestDatasourceService_TestConnection_ClickHouse_PingFails(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	err := svc.TestConnection(context.Background(), &dto.TestConnectionRequest{
		Type: "clickhouse",
		Config: map[string]interface{}{
			"host":     "127.0.0.1",
			"port":     float64(1),
			"username": "u",
			"password": "p",
			"database": "d",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ping failed")
}

func TestDatasourceService_TestConnection_SQLite_PingFails(t *testing.T) {
	// Point sqlite at a directory that cannot be opened as a database file.
	// sql.Open succeeds (lazy), but the first query/ping errors out.
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	err := svc.TestConnection(context.Background(), &dto.TestConnectionRequest{
		Type: "sqlite",
		Config: map[string]interface{}{
			"database": "/nonexistent/dir/that/should/not/exist/db.sqlite",
		},
	})
	assert.Error(t, err)
}

func TestDatasourceService_TestConnection_Excel_MissingFilePath(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	err := svc.TestConnection(context.Background(), &dto.TestConnectionRequest{
		Type:   "excel",
		Config: map[string]interface{}{},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing file_path")
}

func TestDatasourceService_TestConnection_CSV_MissingFilePath(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	err := svc.TestConnection(context.Background(), &dto.TestConnectionRequest{
		Type:   "csv",
		Config: map[string]interface{}{},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing file_path")
}

func TestDatasourceService_Delete_NotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	err := svc.Delete(context.Background(), 99)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasourceService_Update_NotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasourceRepository(db)
	svc := NewDatasourceService(repo, nil)

	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	err := svc.Update(context.Background(), 99, &dto.UpdateDatasourceRequest{Name: "x"})
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
