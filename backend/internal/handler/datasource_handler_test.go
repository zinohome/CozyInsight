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

func setupDatasourceHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewDatasourceRepository(sqlxDB)
	svc := service.NewDatasourceService(repo)
	h := NewDatasourceHandler(svc)

	r := gin.New()
	r.POST("/datasource", h.Create)
	r.GET("/datasource", h.List)
	r.GET("/datasource/:id", h.Get)
	r.PUT("/datasource/:id", h.Update)
	r.DELETE("/datasource/:id", h.Delete)

	return r, mock
}

func TestDatasourceHandler_Create(t *testing.T) {
	r, mock := setupDatasourceHandler(t)

	mock.ExpectExec("INSERT INTO datasources").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.CreateDatasourceRequest{
		Name:   "MySQL Prod",
		Type:   "mysql",
		Config: `{"host":"127.0.0.1"}`,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/datasource", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasourceHandler_Get(t *testing.T) {
	r, mock := setupDatasourceHandler(t)

	columns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT (.+) FROM datasources WHERE id = (.+) AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "MySQL", "mysql", "{}", 1, 1, now, now, nil,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/datasource/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasourceHandler_Get_InvalidID(t *testing.T) {
	r, _ := setupDatasourceHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/datasource/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasourceHandler_List(t *testing.T) {
	r, mock := setupDatasourceHandler(t)

	columns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT (.+) FROM datasources WHERE deleted_at IS NULL").
		WillReturnRows(sqlmock.NewRows(columns))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/datasource", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasourceHandler_Delete(t *testing.T) {
	r, mock := setupDatasourceHandler(t)

	mock.ExpectExec("UPDATE datasources SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/datasource/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasourceHandler_Create_InvalidBody(t *testing.T) {
	r, _ := setupDatasourceHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/datasource", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasourceHandler_Get_NotFound(t *testing.T) {
	r, mock := setupDatasourceHandler(t)

	mock.ExpectQuery("SELECT (.+) FROM datasources WHERE id = (.+) AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/datasource/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDatasourceHandler_List_DBError(t *testing.T) {
	r, mock := setupDatasourceHandler(t)

	mock.ExpectQuery("SELECT (.+) FROM datasources WHERE deleted_at IS NULL").
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/datasource", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDatasourceHandler_Update_Success(t *testing.T) {
	r, mock := setupDatasourceHandler(t)

	columns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT (.+) FROM datasources WHERE id = (.+) AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "MySQL", "mysql", "{}", 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE datasources SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body, _ := json.Marshal(dto.UpdateDatasourceRequest{Name: "Updated"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/datasource/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasourceHandler_Update_InvalidBody(t *testing.T) {
	r, _ := setupDatasourceHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/datasource/1", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasourceHandler_Update_NotFound(t *testing.T) {
	r, mock := setupDatasourceHandler(t)

	mock.ExpectQuery("SELECT (.+) FROM datasources WHERE id = (.+) AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	body, _ := json.Marshal(dto.UpdateDatasourceRequest{Name: "Updated"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/datasource/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasourceHandler_Delete_NotFound(t *testing.T) {
	r, mock := setupDatasourceHandler(t)

	mock.ExpectExec("UPDATE datasources SET deleted_at = NOW").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/datasource/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
