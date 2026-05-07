package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
)

func setupRoleHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewRoleRepository(sqlxDB)
	svc := service.NewRoleService(repo)
	h := NewRoleHandler(svc)

	r := gin.New()
	r.POST("/role", h.Create)
	r.GET("/role", h.List)
	r.GET("/role/:id", h.Get)
	r.PUT("/role/:id", h.Update)
	r.DELETE("/role/:id", h.Delete)
	r.GET("/role/menus", h.ListMenus)
	r.PUT("/role/:id/menus", h.SetRoleMenus)
	r.GET("/role/:id/menus", h.GetRoleMenus)
	r.PUT("/user/:id/roles", h.SetUserRoles)
	return r, mock
}

func TestRoleHandler_Create(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectExec("INSERT INTO roles").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.CreateRoleRequest{
		Name:   "Admin",
		Code:   "admin",
		Status: 1,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_Get(t *testing.T) {
	r, mock := setupRoleHandler(t)

	columns := []string{"id", "name", "code", "description", "status", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM roles WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Admin", "admin", "Administrator", 1, now, now, nil,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/role/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_Get_InvalidID(t *testing.T) {
	r, _ := setupRoleHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/role/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_List(t *testing.T) {
	r, mock := setupRoleHandler(t)

	columns := []string{"id", "name", "code", "description", "status", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM roles WHERE deleted_at IS NULL").
		WillReturnRows(sqlmock.NewRows(columns))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/role", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_Update(t *testing.T) {
	r, mock := setupRoleHandler(t)

	columns := []string{"id", "name", "code", "description", "status", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM roles WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Admin", "admin", "Old desc", 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE roles SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body, _ := json.Marshal(dto.UpdateRoleRequest{
		Name: "Super Admin",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/role/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_Delete(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectExec("UPDATE roles SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/role/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_ListMenus(t *testing.T) {
	r, mock := setupRoleHandler(t)

	columns := []string{"id", "parent_id", "name", "path", "component", "icon", "sort", "status", "created_at", "updated_at"}
	mock.ExpectQuery("SELECT \\* FROM menus WHERE status = 1").
		WillReturnRows(sqlmock.NewRows(columns))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/role/menus", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_SetRoleMenus(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectExec("DELETE FROM role_menus WHERE role_id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("INSERT INTO role_menus").
		WillReturnResult(sqlmock.NewResult(0, 2))

	body, _ := json.Marshal(dto.SetRoleMenusRequest{
		MenuIDs: []uint64{1, 2},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/role/1/menus", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_GetRoleMenus(t *testing.T) {
	r, mock := setupRoleHandler(t)

	columns := []string{"menu_id"}
	mock.ExpectQuery("SELECT menu_id FROM role_menus WHERE role_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1).AddRow(2))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/role/1/menus", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_SetUserRoles(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectExec("DELETE FROM user_roles WHERE user_id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("INSERT INTO user_roles").
		WillReturnResult(sqlmock.NewResult(0, 2))

	body, _ := json.Marshal(dto.SetUserRolesRequest{
		RoleIDs: []uint64{1, 2},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/user/1/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_Create_InvalidBody(t *testing.T) {
	r, _ := setupRoleHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/role", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_Get_NotFound(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectQuery("SELECT \\* FROM roles WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/role/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRoleHandler_List_DBError(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectQuery("SELECT \\* FROM roles WHERE deleted_at IS NULL").
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/role", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRoleHandler_Update_NotFound(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectQuery("SELECT \\* FROM roles WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	body, _ := json.Marshal(dto.UpdateRoleRequest{Name: "Updated"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/role/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_Update_InvalidBody(t *testing.T) {
	r, _ := setupRoleHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/role/1", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_Delete_NotFound(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectExec("UPDATE roles SET deleted_at = NOW").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/role/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_ListMenus_DBError(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectQuery("SELECT \\* FROM menus WHERE status = 1").
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/role/menus", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRoleHandler_SetRoleMenus_InvalidBody(t *testing.T) {
	r, _ := setupRoleHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/role/1/menus", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleHandler_GetRoleMenus_DBError(t *testing.T) {
	r, mock := setupRoleHandler(t)

	mock.ExpectQuery("SELECT menu_id FROM role_menus WHERE role_id = \\?").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/role/1/menus", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
