package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/repository"
	"cozy-insight/internal/testutil"
)

func TestMessageService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo)

	mock.ExpectExec("INSERT INTO messages").
		WillReturnResult(sqlmock.NewResult(1, 1))

	msg, err := svc.Create(context.Background(), 7, "Welcome", "Hello world", "system")
	require.NoError(t, err)
	assert.Equal(t, uint64(7), msg.UserID)
	assert.Equal(t, "Welcome", msg.Title)
	assert.Equal(t, "Hello world", msg.Content)
	assert.Equal(t, "system", msg.Type)
	assert.Equal(t, int8(0), msg.IsRead)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageService_List_All(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo)

	now := time.Now()
	mock.ExpectQuery("FROM messages WHERE user_id = \\? ORDER BY created_at DESC").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "content", "type", "is_read", "created_at"}).
			AddRow(1, 7, "A", "a", "info", 0, now).
			AddRow(2, 7, "B", "b", "warn", 1, now))

	list, err := svc.List(context.Background(), 7, false)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageService_List_UnreadOnly(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo)

	mock.ExpectQuery("AND is_read = 0").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "content", "type", "is_read", "created_at"}))

	_, err := svc.List(context.Background(), 7, true)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageService_MarkAsRead(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo)

	mock.ExpectExec("UPDATE messages SET is_read = 1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.MarkAsRead(context.Background(), 1, 7)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageService_MarkAllAsRead(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo)

	mock.ExpectExec("UPDATE messages SET is_read = 1 WHERE user_id = \\? AND is_read = 0").
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := svc.MarkAllAsRead(context.Background(), 7)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageService_CountUnread(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(5))

	n, err := svc.CountUnread(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo)

	mock.ExpectExec("DELETE FROM messages").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1, 7)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageService_Create_DBError(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo)

	mock.ExpectExec("INSERT INTO messages").
		WillReturnError(sql.ErrConnDone)

	_, err := svc.Create(context.Background(), 7, "x", "y", "info")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create message failed")
}
