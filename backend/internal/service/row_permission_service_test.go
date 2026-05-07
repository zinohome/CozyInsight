package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/repository"
	"cozy-insight/internal/testutil"
)

func TestRowPermissionService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRowPermissionRepository(db)
	svc := NewRowPermissionService(repo)

	mock.ExpectExec("INSERT INTO row_permissions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), 1, "dept_id", "=", "1", "dept_id")
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "dept_id", result.FieldName)
	assert.Equal(t, "=", result.Operator)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRowPermissionService_ListByDataset(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRowPermissionRepository(db)
	svc := NewRowPermissionService(repo)

	columns := []string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, 1, "dept_id", "=", "1", "dept_id", 1, now, now))

	list, err := svc.ListByDataset(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "dept_id", list[0].FieldName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRowPermissionService_BuildRowFilter(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRowPermissionRepository(db)
	svc := NewRowPermissionService(repo)

	columns := []string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, 1, "dept_id", "=", "1", "dept_id", 1, now, now).
			AddRow(2, 1, "region", "LIKE", "east", "region", 1, now, now))

	conditions, err := svc.BuildRowFilter(context.Background(), 1, map[string]string{
		"dept_id": "5",
	})
	require.NoError(t, err)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "dept_id", conditions[0].FieldName)
	assert.Equal(t, "=", conditions[0].Operator)
	assert.Equal(t, "5", conditions[0].Value)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRowPermissionService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRowPermissionRepository(db)
	svc := NewRowPermissionService(repo)

	mock.ExpectExec("DELETE FROM row_permissions WHERE id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRowPermissionService_BuildRowFilter_NoPermissions(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRowPermissionRepository(db)
	svc := NewRowPermissionService(repo)

	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}))

	conditions, err := svc.BuildRowFilter(context.Background(), 1, map[string]string{
		"dept_id": "5",
	})
	require.NoError(t, err)
	assert.Len(t, conditions, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRowPermissionService_BuildRowFilter_NoMatchingAttr(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRowPermissionRepository(db)
	svc := NewRowPermissionService(repo)

	columns := []string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, 1, "dept_id", "=", "1", "dept_id", 1, now, now))

	// userAttrs does not contain "dept_id"
	conditions, err := svc.BuildRowFilter(context.Background(), 1, map[string]string{})
	require.NoError(t, err)
	assert.Len(t, conditions, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRowPermissionService_BuildRowFilter_InvalidOperator(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRowPermissionRepository(db)
	svc := NewRowPermissionService(repo)

	columns := []string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, 1, "dept_id", "DROP", "1", "dept_id", 1, now, now))

	conditions, err := svc.BuildRowFilter(context.Background(), 1, map[string]string{
		"dept_id": "5",
	})
	require.NoError(t, err)
	assert.Len(t, conditions, 0) // DROP is not in allowedOperators
	assert.NoError(t, mock.ExpectationsWereMet())
}
