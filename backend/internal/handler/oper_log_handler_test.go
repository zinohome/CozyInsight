package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
)

func setupOperationLogHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewOperationLogRepository(sqlxDB)
	svc := service.NewOperationLogService(repo)
	h := NewOperationLogHandler(svc)

	r := gin.New()
	r.GET("/operation-logs", h.List)
	return r, mock
}

func TestOperationLogHandler_List(t *testing.T) {
	r, mock := setupOperationLogHandler(t)

	columns := []string{"id", "user_id", "username", "method", "path", "query", "body", "ip", "user_agent", "status_code", "duration", "error_message", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM operation_logs ORDER BY created_at DESC LIMIT").
		WillReturnRows(sqlmock.NewRows(columns))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/operation-logs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationLogHandler_List_DBError(t *testing.T) {
	r, mock := setupOperationLogHandler(t)

	mock.ExpectQuery("SELECT \\* FROM operation_logs ORDER BY created_at DESC LIMIT").
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/operation-logs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
