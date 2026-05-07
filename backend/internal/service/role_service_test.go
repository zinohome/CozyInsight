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

func TestRoleService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	mock.ExpectExec("INSERT INTO roles").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateRoleRequest{
		Name: "Admin",
		Code: "admin",
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "Admin", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleService_GetByID(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	columns := []string{"id", "name", "code", "description", "status", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM roles WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Admin", "admin", "Admin role", 1, now, now, nil,
		))

	role, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), role.ID)
	assert.Equal(t, "Admin", role.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleService_List(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	columns := []string{"id", "name", "code", "description", "status", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM roles WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "Admin", "admin", "", 1, now, now, nil).
			AddRow(2, "User", "user", "", 1, now, now, nil))

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "Admin", list[0].Name)
	assert.Equal(t, "User", list[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleService_Update(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	columns := []string{"id", "name", "code", "description", "status", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM roles WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Admin", "admin", "Old desc", 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE roles SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	status := int8(0)
	err := svc.Update(context.Background(), 1, &dto.UpdateRoleRequest{
		Name:        "SuperAdmin",
		Code:        "superadmin",
		Description: "Updated desc",
		Status:      &status,
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleService_Update_NotFound(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	mock.ExpectQuery("SELECT \\* FROM roles WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	err := svc.Update(context.Background(), 1, &dto.UpdateRoleRequest{Name: "test"})
	assert.Error(t, err)
}

func TestRoleService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	mock.ExpectExec("UPDATE roles SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleService_ListMenus(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	columns := []string{"id", "parent_id", "name", "path", "component", "icon", "sort", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM menus WHERE status = 1 ORDER BY parent_id, sort").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, 0, "System", "/system", "", "setting", 1, 1, now, now,
		))

	list, err := svc.ListMenus(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "System", list[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleService_SetRoleMenus(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	mock.ExpectExec("DELETE FROM role_menus WHERE role_id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("INSERT INTO role_menus").
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := svc.SetRoleMenus(context.Background(), 1, []uint64{1, 2})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleService_GetRoleMenus(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	mock.ExpectQuery("SELECT menu_id FROM role_menus WHERE role_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"menu_id"}).AddRow(1).AddRow(2))

	ids, err := svc.GetRoleMenus(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []uint64{1, 2}, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleService_SetUserRoles(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	mock.ExpectExec("DELETE FROM user_roles WHERE user_id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("INSERT INTO user_roles").
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := svc.SetUserRoles(context.Background(), 1, []uint64{1, 2})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleService_GetUserRoles(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewRoleRepository(db)
	svc := NewRoleService(repo)

	mock.ExpectQuery("SELECT role_id FROM user_roles WHERE user_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"role_id"}).AddRow(1).AddRow(2))

	ids, err := svc.GetUserRoles(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []uint64{1, 2}, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}
