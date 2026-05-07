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

func setupChartHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewChartRepository(sqlxDB)
	svc := service.NewChartService(repo)
	h := NewChartHandler(svc)

	r := gin.New()
	r.POST("/chart", h.Create)
	r.GET("/chart", h.List)
	r.GET("/chart/:id", h.Get)
	r.PUT("/chart/:id", h.Update)
	r.DELETE("/chart/:id", h.Delete)
	return r, mock
}

func TestChartHandler_Create(t *testing.T) {
	r, mock := setupChartHandler(t)

	mock.ExpectExec("INSERT INTO charts").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.CreateChartRequest{
		Title:     "Test Chart",
		Type:      "bar",
		DatasetID: 1,
		Config:    "{}",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/chart", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChartHandler_Get(t *testing.T) {
	r, mock := setupChartHandler(t)

	columns := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test Chart", "bar", 1, "{}", 1, 1, now, now, nil,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chart/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChartHandler_Get_InvalidID(t *testing.T) {
	r, _ := setupChartHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chart/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChartHandler_List(t *testing.T) {
	r, mock := setupChartHandler(t)

	columns := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM charts WHERE deleted_at IS NULL").
		WillReturnRows(sqlmock.NewRows(columns))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chart", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChartHandler_Update(t *testing.T) {
	r, mock := setupChartHandler(t)

	columns := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old Title", "bar", 1, "{}", 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE charts SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body, _ := json.Marshal(dto.UpdateChartRequest{
		Title: "Updated Chart",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/chart/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChartHandler_Delete(t *testing.T) {
	r, mock := setupChartHandler(t)

	mock.ExpectExec("UPDATE charts SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/chart/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChartHandler_Create_InvalidBody(t *testing.T) {
	r, _ := setupChartHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/chart", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChartHandler_Get_NotFound(t *testing.T) {
	r, mock := setupChartHandler(t)

	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chart/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChartHandler_List_DBError(t *testing.T) {
	r, mock := setupChartHandler(t)

	mock.ExpectQuery("SELECT \\* FROM charts WHERE deleted_at IS NULL").
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chart", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestChartHandler_Update_NotFound(t *testing.T) {
	r, mock := setupChartHandler(t)

	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	body, _ := json.Marshal(dto.UpdateChartRequest{Title: "Updated"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/chart/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChartHandler_Update_InvalidBody(t *testing.T) {
	r, _ := setupChartHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/chart/1", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChartHandler_Delete_NotFound(t *testing.T) {
	r, mock := setupChartHandler(t)

	mock.ExpectExec("UPDATE charts SET deleted_at = NOW").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/chart/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
