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

func setupMessageHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewMessageRepository(sqlxDB)
	svc := service.NewMessageService(repo)
	h := NewMessageHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint64(1))
		c.Next()
	})
	r.GET("/messages", h.List)
	r.GET("/messages/unread-count", h.CountUnread)
	r.POST("/messages/:id/read", h.MarkAsRead)
	r.POST("/messages/read-all", h.MarkAllAsRead)
	r.DELETE("/messages/:id", h.Delete)
	return r, mock
}

func TestMessageHandler_List(t *testing.T) {
	r, mock := setupMessageHandler(t)

	cols := []string{"id", "user_id", "title", "content", "type", "is_read", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM messages WHERE user_id = \\? ORDER BY created_at DESC").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, 1, "hi", "yo", "info", 0, now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/messages", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_List_UnreadOnly(t *testing.T) {
	r, mock := setupMessageHandler(t)

	cols := []string{"id", "user_id", "title", "content", "type", "is_read", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM messages WHERE user_id = \\? AND is_read = 0 ORDER BY created_at DESC").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, 1, "hi", "yo", "info", 0, now))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/messages?unreadOnly=true", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_List_DBError(t *testing.T) {
	r, mock := setupMessageHandler(t)

	mock.ExpectQuery("SELECT \\* FROM messages WHERE user_id = \\? ORDER BY created_at DESC").
		WithArgs(uint64(1)).
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/messages", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_CountUnread(t *testing.T) {
	r, mock := setupMessageHandler(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(7)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM messages WHERE user_id = \\? AND is_read = 0").
		WithArgs(uint64(1)).
		WillReturnRows(rows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/messages/unread-count", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_CountUnread_DBError(t *testing.T) {
	r, mock := setupMessageHandler(t)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM messages WHERE user_id = \\? AND is_read = 0").
		WithArgs(uint64(1)).
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/messages/unread-count", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_MarkAsRead(t *testing.T) {
	r, mock := setupMessageHandler(t)

	mock.ExpectExec("UPDATE messages SET is_read = 1 WHERE id = \\?").
		WithArgs(uint64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages/7/read", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_MarkAsRead_InvalidID(t *testing.T) {
	r, _ := setupMessageHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages/abc/read", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMessageHandler_MarkAsRead_DBError(t *testing.T) {
	r, mock := setupMessageHandler(t)

	mock.ExpectExec("UPDATE messages SET is_read = 1 WHERE id = \\?").
		WithArgs(uint64(7)).
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages/7/read", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_MarkAllAsRead(t *testing.T) {
	r, mock := setupMessageHandler(t)

	mock.ExpectExec("UPDATE messages SET is_read = 1 WHERE user_id = \\? AND is_read = 0").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 3))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages/read-all", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_MarkAllAsRead_DBError(t *testing.T) {
	r, mock := setupMessageHandler(t)

	mock.ExpectExec("UPDATE messages SET is_read = 1 WHERE user_id = \\? AND is_read = 0").
		WithArgs(uint64(1)).
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/messages/read-all", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_Delete(t *testing.T) {
	r, mock := setupMessageHandler(t)

	mock.ExpectExec("DELETE FROM messages WHERE id = \\?").
		WithArgs(uint64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/messages/7", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageHandler_Delete_InvalidID(t *testing.T) {
	r, _ := setupMessageHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/messages/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMessageHandler_Delete_DBError(t *testing.T) {
	r, mock := setupMessageHandler(t)

	mock.ExpectExec("DELETE FROM messages WHERE id = \\?").
		WithArgs(uint64(7)).
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/messages/7", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
