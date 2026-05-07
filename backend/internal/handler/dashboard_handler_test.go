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

func setupDashboardHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewDashboardRepository(sqlxDB)
	svc := service.NewDashboardService(repo)
	h := NewDashboardHandler(svc)

	r := gin.New()
	r.POST("/dashboard", h.Create)
	r.GET("/dashboard", h.List)
	r.GET("/dashboard/:id", h.Get)
	r.PUT("/dashboard/:id", h.Update)
	r.DELETE("/dashboard/:id", h.Delete)
	r.POST("/dashboard/:id/charts", h.AddChart)
	r.GET("/dashboard/:id/charts", h.GetCharts)
	r.DELETE("/dashboard/:id/charts/:chartId", h.RemoveChart)
	return r, mock
}

func TestDashboardHandler_Create(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectExec("INSERT INTO dashboards").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.CreateDashboardRequest{
		Title:  "Test Dashboard",
		Config: "{}",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dashboard", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_Get(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	columns := []string{"id", "title", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test Dashboard", "{}", 1, 1, now, now, nil,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_Get_InvalidID(t *testing.T) {
	r, _ := setupDashboardHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_List(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	columns := []string{"id", "title", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE deleted_at IS NULL").
		WillReturnRows(sqlmock.NewRows(columns))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_Update(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	columns := []string{"id", "title", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old Dashboard", "{}", 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE dashboards SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body, _ := json.Marshal(dto.UpdateDashboardRequest{
		Title: "Updated Dashboard",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/dashboard/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_Delete(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectExec("UPDATE dashboards SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/dashboard/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_AddChart(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectExec("INSERT INTO dashboard_charts").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.AddChartToDashboardRequest{
		ChartID:   1,
		PositionX: 0,
		PositionY: 0,
		Width:     6,
		Height:    4,
		Config:    "{}",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dashboard/1/charts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_GetCharts(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	columns := []string{"id", "dashboard_id", "chart_id", "position_x", "position_y", "width", "height", "config", "created_at", "updated_at"}
	mock.ExpectQuery("SELECT \\* FROM dashboard_charts WHERE dashboard_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/1/charts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_RemoveChart(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectExec("DELETE FROM dashboard_charts WHERE dashboard_id = \\? AND chart_id = \\?").
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/dashboard/1/charts/2", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_Create_InvalidBody(t *testing.T) {
	r, _ := setupDashboardHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dashboard", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_Get_NotFound(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDashboardHandler_List_DBError(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE deleted_at IS NULL").
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDashboardHandler_Update_NotFound(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	body, _ := json.Marshal(dto.UpdateDashboardRequest{Title: "Updated"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/dashboard/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_Update_InvalidBody(t *testing.T) {
	r, _ := setupDashboardHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/dashboard/1", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_Delete_NotFound(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectExec("UPDATE dashboards SET deleted_at = NOW").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/dashboard/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_GetCharts_DBError(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectQuery("SELECT \\* FROM dashboard_charts WHERE dashboard_id = \\?").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard/1/charts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDashboardHandler_RemoveChart_NotFound(t *testing.T) {
	r, mock := setupDashboardHandler(t)

	mock.ExpectExec("DELETE FROM dashboard_charts WHERE dashboard_id = \\? AND chart_id = \\?").
		WithArgs(1, 2).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/dashboard/1/charts/2", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_AddChart_InvalidBody(t *testing.T) {
	r, _ := setupDashboardHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dashboard/1/charts", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
