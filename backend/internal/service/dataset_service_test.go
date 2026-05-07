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

func TestDatasetService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil)

	mock.ExpectExec("INSERT INTO datasets").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateDatasetRequest{
		Name:         "Test Dataset",
		DatasourceID: 1,
		TableName:    "users",
		Type:         "table",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "Test Dataset", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_GetByID(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test", 1, "db", "users", "table", 0, 1, 1, now, now, nil,
		))

	ds, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), ds.ID)
	assert.Equal(t, "Test", ds.Name)
	assert.Equal(t, "users", ds.TableName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_inferDeType(t *testing.T) {
	svc := NewDatasetService(nil, nil, nil)

	assert.Equal(t, int8(2), svc.inferDeType("BIGINT"))
	assert.Equal(t, int8(0), svc.inferDeType("VARCHAR"))
	assert.Equal(t, int8(1), svc.inferDeType("DATETIME"))
	assert.Equal(t, int8(2), svc.inferDeType("INT"))
	assert.Equal(t, int8(2), svc.inferDeType("FLOAT"))
	assert.Equal(t, int8(2), svc.inferDeType("DOUBLE"))
	assert.Equal(t, int8(2), svc.inferDeType("DECIMAL"))
	assert.Equal(t, int8(1), svc.inferDeType("DATE"))
	assert.Equal(t, int8(1), svc.inferDeType("TIME"))
	assert.Equal(t, int8(0), svc.inferDeType("TEXT"))
	assert.Equal(t, int8(0), svc.inferDeType("CHAR"))
	assert.Equal(t, int8(4), svc.inferDeType("UNKNOWN"))
}

func TestDatasetService_List(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test", 1, "db", "users", "table", 0, 1, 1, now, now, nil,
		))

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_Update(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old", 1, "db", "users", "table", 0, 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE datasets SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mode := int8(1)
	status := int8(0)
	dsid := uint64(2)
	err := svc.Update(context.Background(), 1, &dto.UpdateDatasetRequest{
		Name:         "Updated",
		DatasourceID: &dsid,
		DatabaseName: "newdb",
		TableName:    "orders",
		Type:         "sql",
		Mode:         &mode,
		Status:       &status,
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	svc := NewDatasetService(repo, nil, nil)

	mock.ExpectExec("UPDATE datasets SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatasetService_getTableColumns(t *testing.T) {
	svc := NewDatasetService(nil, nil, nil)
	cols, err := svc.getTableColumns(context.Background(), nil, "db", "users")
	require.NoError(t, err)
	assert.Len(t, cols, 3)
	assert.Equal(t, "id", cols[0].Name)
}

func TestDatasetService_queryTableData(t *testing.T) {
	svc := NewDatasetService(nil, nil, nil)
	data, err := svc.queryTableData(context.Background(), nil, "db", "users", 10)
	require.NoError(t, err)
	assert.Len(t, data, 2)
	assert.Equal(t, "Test 1", data[0]["name"])
}

func TestDatasetService_buildRowFilter(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	rowPermRepo := repository.NewRowPermissionRepository(db)
	svc := NewDatasetService(repo, nil, rowPermRepo)

	columns := []string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, 1, "dept_id", "=", "1", "dept_id", 1, now, now,
		))

	conditions, err := svc.buildRowFilter(context.Background(), 1, map[string]string{"dept_id": "5"})
	require.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "dept_id", conditions[0].FieldName)
	assert.Equal(t, "=", conditions[0].Operator)
	assert.Equal(t, "5", conditions[0].Value)
}
