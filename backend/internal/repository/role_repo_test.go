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

func TestRoleRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO roles").
		WillReturnResult(sqlmock.NewResult(11, 1))

	r := &model.Role{Name: "admin", Code: "ADMIN", Description: "Administrator", Status: 1}
	require.NoError(t, repo.Create(context.Background(), r))
	assert.Equal(t, uint64(11), r.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "name", "code", "description", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM roles WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "admin", "ADMIN", "desc", 1, now, now))

	r, err := repo.FindByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "admin", r.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "name", "code", "description", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM roles WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "admin", "ADMIN", "", 1, now, now).
			AddRow(2, "user", "USER", "", 1, now, now))

	rs, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, rs, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE roles SET name").
		WillReturnResult(sqlmock.NewResult(0, 1))

	r := &model.Role{ID: 1, Name: "newname", Code: "ADMIN", Description: "x", Status: 1}
	require.NoError(t, repo.Update(context.Background(), r))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE roles SET deleted_at").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_ListMenus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "parent_id", "name", "path", "component", "icon", "sort", "status", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM menus WHERE status = 1 ORDER BY parent_id, sort").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, 0, "Dashboard", "/dashboard", "Dashboard", "icon", 1, 1, now, now))

	menus, err := repo.ListMenus(context.Background())
	require.NoError(t, err)
	assert.Len(t, menus, 1)
	assert.Equal(t, "Dashboard", menus[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_SetRoleMenus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("DELETE FROM role_menus WHERE role_id = \\?").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO role_menus \\(role_id, menu_id\\) VALUES").
		WillReturnResult(sqlmock.NewResult(0, 2))

	require.NoError(t, repo.SetRoleMenus(context.Background(), 1, []uint64{10, 20}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_SetRoleMenus_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	// only DELETE; no INSERT when menuIDs is empty
	mock.ExpectExec("DELETE FROM role_menus WHERE role_id = \\?").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.SetRoleMenus(context.Background(), 1, []uint64{}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_GetRoleMenus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectQuery("SELECT menu_id FROM role_menus WHERE role_id = \\?").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"menu_id"}).AddRow(10).AddRow(20))

	ids, err := repo.GetRoleMenus(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []uint64{10, 20}, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_SetUserRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("DELETE FROM user_roles WHERE user_id = \\?").
		WithArgs(uint64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO user_roles \\(user_id, role_id\\) VALUES").
		WillReturnResult(sqlmock.NewResult(0, 2))

	require.NoError(t, repo.SetUserRoles(context.Background(), 5, []uint64{1, 2}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_SetUserRoles_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("DELETE FROM user_roles WHERE user_id = \\?").
		WithArgs(uint64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.SetUserRoles(context.Background(), 5, nil))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_GetUserRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRoleRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectQuery("SELECT role_id FROM user_roles WHERE user_id = \\?").
		WithArgs(uint64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"role_id"}).AddRow(1).AddRow(2))

	ids, err := repo.GetUserRoles(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, []uint64{1, 2}, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}
