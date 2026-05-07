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

	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
)

func setupRowPermissionHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewRowPermissionRepository(sqlxDB)
	svc := service.NewRowPermissionService(repo)
	h := NewRowPermissionHandler(svc)

	r := gin.New()
	r.GET("/dataset/:id/row-permissions", h.List)
	r.POST("/dataset/:id/row-permissions", h.Create)
	r.DELETE("/row-permissions/:permId", h.Delete)
	return r, mock
}

func TestRowPermissionHandler_List(t *testing.T) {
	r, mock := setupRowPermissionHandler(t)

	columns := []string{"id", "dataset_id", "field_name", "operator", "value", "user_attr", "status", "created_at", "updated_at"}
	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, 1, "dept_id", "=", "1", "dept", 1, time.Now(), time.Now(),
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset/1/row-permissions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRowPermissionHandler_List_InvalidID(t *testing.T) {
	r, _ := setupRowPermissionHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset/abc/row-permissions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRowPermissionHandler_Create(t *testing.T) {
	r, mock := setupRowPermissionHandler(t)

	mock.ExpectExec("INSERT INTO row_permissions").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(map[string]string{
		"fieldName": "dept_id",
		"operator":  "=",
		"value":     "1",
		"userAttr":  "dept",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dataset/1/row-permissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRowPermissionHandler_Create_InvalidID(t *testing.T) {
	r, _ := setupRowPermissionHandler(t)

	body, _ := json.Marshal(map[string]string{
		"fieldName": "dept_id",
		"operator":  "=",
		"value":     "1",
		"userAttr":  "dept",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dataset/abc/row-permissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRowPermissionHandler_Delete(t *testing.T) {
	r, mock := setupRowPermissionHandler(t)

	mock.ExpectExec("DELETE FROM row_permissions WHERE id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/row-permissions/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRowPermissionHandler_Delete_InvalidID(t *testing.T) {
	r, _ := setupRowPermissionHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/row-permissions/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRowPermissionHandler_Create_InvalidBody(t *testing.T) {
	r, _ := setupRowPermissionHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dataset/1/row-permissions", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRowPermissionHandler_List_DBError(t *testing.T) {
	r, mock := setupRowPermissionHandler(t)

	mock.ExpectQuery("SELECT \\* FROM row_permissions WHERE dataset_id = \\? AND status = 1").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset/1/row-permissions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRowPermissionHandler_Delete_NotFound(t *testing.T) {
	r, mock := setupRowPermissionHandler(t)

	mock.ExpectExec("DELETE FROM row_permissions WHERE id = \\?").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/row-permissions/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
