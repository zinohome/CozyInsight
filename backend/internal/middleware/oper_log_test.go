package middleware

import (
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
)

func TestOperationLog_SkipsAuthPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewOperationLogRepository(sqlxDB)

	r := gin.New()
	r.Use(OperationLog(repo))
	r.GET("/health", func(c *gin.Context) { c.Status(200) })
	r.POST("/api/v1/auth/login", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// no DB calls expected for skipped paths
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLog_RecordsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewOperationLogRepository(sqlxDB)

	mock.ExpectExec("INSERT INTO operation_logs").
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, uint64(1))
		c.Set(ContextKeyUsername, "tester")
		c.Next()
	})
	r.Use(OperationLog(repo))
	r.GET("/api/v1/dataset", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/dataset", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// wait for goroutine
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, mock.ExpectationsWereMet())
}
