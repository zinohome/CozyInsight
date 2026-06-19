package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

func setupScheduleHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewScheduleTaskRepository(sqlxDB)
	svc := service.NewScheduleService(repo)
	h := NewScheduleHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint64(1))
		c.Next()
	})
	r.GET("/schedule", h.List)
	r.POST("/schedule", h.Create)
	r.GET("/schedule/:id", h.Get)
	r.PUT("/schedule/:id", h.Update)
	r.DELETE("/schedule/:id", h.Delete)
	r.POST("/schedule/:id/enable", h.Enable)
	r.POST("/schedule/:id/disable", h.Disable)
	r.POST("/schedule/:id/execute", h.Execute)
	return r, mock
}

func scheduleColumns() []string {
	return []string{"id", "name", "type", "cron_expr", "config", "enabled", "status", "created_by", "created_at", "update_time"}
}

func scheduleRow(id uint64, taskType string) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(scheduleColumns()).AddRow(
		id, "task"+strconv.FormatUint(id, 10), taskType, "0 9 * * *",
		"{}", true, "inactive", uint64(1), now, now,
	)
}

func TestScheduleHandler_Create(t *testing.T) {
	r, mock := setupScheduleHandler(t)

	mock.ExpectExec("INSERT INTO schedule_tasks").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"name":"daily-report","type":"email_report","cronExpr":"0 9 * * *","config":"{}"}`
	req, _ := http.NewRequest("POST", "/schedule", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleHandler_Create_BadJSON(t *testing.T) {
	r, _ := setupScheduleHandler(t)

	req, _ := http.NewRequest("POST", "/schedule", strings.NewReader("{not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScheduleHandler_Get(t *testing.T) {
	r, mock := setupScheduleHandler(t)

	mock.ExpectQuery("SELECT \\* FROM schedule_tasks WHERE id = \\?").
		WithArgs(uint64(7)).
		WillReturnRows(scheduleRow(7, "email_report"))

	req, _ := http.NewRequest("GET", "/schedule/7", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleHandler_Get_InvalidID(t *testing.T) {
	r, _ := setupScheduleHandler(t)

	req, _ := http.NewRequest("GET", "/schedule/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScheduleHandler_Delete(t *testing.T) {
	r, mock := setupScheduleHandler(t)

	mock.ExpectExec("DELETE FROM schedule_tasks WHERE id = \\?").
		WithArgs(uint64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest("DELETE", "/schedule/5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleHandler_List(t *testing.T) {
	r, mock := setupScheduleHandler(t)

	mock.ExpectQuery("SELECT \\* FROM schedule_tasks ORDER BY id DESC").
		WillReturnRows(sqlmock.NewRows(scheduleColumns()).
			AddRow(1, "a", "email_report", "0 9 * * *", "{}", true, "active", 1, time.Now(), time.Now()).
			AddRow(2, "b", "snapshot", "0 0 * * *", "{}", true, "active", 1, time.Now(), time.Now()))

	req, _ := http.NewRequest("GET", "/schedule", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleHandler_Execute_NotFound(t *testing.T) {
	r, mock := setupScheduleHandler(t)

	mock.ExpectQuery("SELECT \\* FROM schedule_tasks WHERE id = \\?").
		WithArgs(uint64(404)).
		WillReturnRows(sqlmock.NewRows(scheduleColumns()))

	req, _ := http.NewRequest("POST", "/schedule/404/execute", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "missing task should return 500")
	assert.NoError(t, mock.ExpectationsWereMet())
}