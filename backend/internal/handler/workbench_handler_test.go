package handler

import (
	"bytes"
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
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
)

func setupWorkbenchHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewWorkbenchRepository(sqlxDB)
	svc := service.NewWorkbenchService(repo)
	h := NewWorkbenchHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, uint64(1))
		c.Next()
	})
	r.GET("/workbench/stats", h.GetStats)
	r.GET("/workbench/recent", h.GetRecentViews)
	r.POST("/workbench/recent", h.RecordVisit)
	r.GET("/workbench/favorites", h.GetFavorites)
	r.POST("/workbench/favorites", h.AddFavorite)
	r.DELETE("/workbench/favorites/:type/:id", h.DeleteFavorite)

	return r, mock
}

func TestWorkbenchHandler_GetStats(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM datasources").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM datasets").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM charts").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM dashboards").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM dashboards").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/workbench/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(3), data["datasourceCount"])
	assert.Equal(t, float64(5), data["datasetCount"])
	assert.Equal(t, float64(8), data["chartCount"])
	assert.Equal(t, float64(2), data["dashboardCount"])
	assert.Equal(t, float64(1), data["screenCount"])
}

func TestWorkbenchHandler_GetRecentViews(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	now := time.Now()
	cols := []string{"id", "title", "type", "visited_at"}
	mock.ExpectQuery("SELECT d.id, d.title, d.type, rv.visited_at").
		WithArgs(1, 20).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, "Sales", "dashboard", now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/workbench/recent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
	item := data[0].(map[string]interface{})
	assert.Equal(t, "Sales", item["title"])
	assert.Equal(t, "dashboard", item["type"])
}

func TestWorkbenchHandler_RecordVisit(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	mock.ExpectExec("INSERT INTO recent_views").
		WithArgs(1, "dashboard", uint64(5)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.RecordVisitRequest{ResourceType: "dashboard", ResourceID: 5})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/workbench/recent", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkbenchHandler_AddFavorite(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	mock.ExpectExec("INSERT INTO favorites").
		WithArgs(1, "dashboard", uint64(10)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.AddFavoriteRequest{ResourceType: "dashboard", ResourceID: 10})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/workbench/favorites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkbenchHandler_DeleteFavorite(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	mock.ExpectExec("DELETE FROM favorites").
		WithArgs(1, "screen", uint64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/workbench/favorites/screen/3", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkbenchHandler_GetFavorites(t *testing.T) {
	r, mock := setupWorkbenchHandler(t)

	now := time.Now()
	cols := []string{"id", "title", "type", "created_at"}
	mock.ExpectQuery("SELECT d.id, d.title, d.type, f.created_at").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(2, "KPI", "dashboard", now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/workbench/favorites", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}
