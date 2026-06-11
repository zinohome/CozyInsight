package handler

import (
	"database/sql"
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

func setupShareHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	shareLinkRepo := repository.NewShareLinkRepository(sqlxDB)
	dashboardRepo := repository.NewDashboardRepository(sqlxDB)
	svc := service.NewShareLinkService(shareLinkRepo, dashboardRepo)
	h := NewShareHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint64(1))
		c.Next()
	})
	r.GET("/share/:token", h.GetDashboard)
	r.GET("/share", h.List)
	return r, mock
}

func TestShareHandler_GetDashboard(t *testing.T) {
	r, mock := setupShareHandler(t)

	now := time.Now()
	slColumns := []string{"id", "token", "resource_type", "resource_id", "created_by", "expire_at", "status", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM share_links WHERE token = \\? AND status = 1").
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows(slColumns).AddRow(
			1, "abc123", "dashboard", 1, 1, nil, 1, now,
		))

	dbColumns := []string{"id", "title", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dbColumns).AddRow(
			1, "Sales Dashboard", "{}", 1, 1, now, now, nil,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/share/abc123", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShareHandler_GetDashboard_NotFound(t *testing.T) {
	r, mock := setupShareHandler(t)

	mock.ExpectQuery("SELECT \\* FROM share_links WHERE token = \\? AND status = 1").
		WithArgs("invalid").
		WillReturnError(sqlmock.ErrCancelled)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/share/invalid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestShareHandler_List(t *testing.T) {
	r, mock := setupShareHandler(t)

	cols := []string{"id", "token", "resource_type", "resource_id", "created_by", "expire_at", "password", "status", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM share_links WHERE created_by = \\? AND status = 1 ORDER BY created_at DESC").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, "abc", "dashboard", 1, 1, nil, "", 1, now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/share", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShareHandler_List_DBError(t *testing.T) {
	r, mock := setupShareHandler(t)

	mock.ExpectQuery("SELECT \\* FROM share_links WHERE created_by = \\? AND status = 1 ORDER BY created_at DESC").
		WithArgs(uint64(1)).
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/share", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
