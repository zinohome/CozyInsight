package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/model"
)

func TestMessageRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMessageRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO messages").
		WillReturnResult(sqlmock.NewResult(4, 1))

	msg := &model.Message{UserID: 1, Title: "hi", Content: "hello", Type: "system", IsRead: 0}
	require.NoError(t, repo.Create(context.Background(), msg))
	assert.Equal(t, uint64(4), msg.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_ListByUser_All(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMessageRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "user_id", "title", "content", "type", "is_read", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM messages WHERE user_id = \\? ORDER BY created_at DESC").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 1, "t", "c", "system", 0, now))

	msgs, err := repo.ListByUser(context.Background(), 1, false)
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_ListByUser_UnreadOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMessageRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "user_id", "title", "content", "type", "is_read", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM messages WHERE user_id = \\? AND is_read = 0 ORDER BY created_at DESC").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 1, "t", "c", "system", 0, now))

	msgs, err := repo.ListByUser(context.Background(), 1, true)
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_MarkAsRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMessageRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE messages SET is_read = 1 WHERE id = \\?").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.MarkAsRead(context.Background(), 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_MarkAllAsRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMessageRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("UPDATE messages SET is_read = 1 WHERE user_id = \\? AND is_read = 0").
		WithArgs(uint64(1)).
		WillReturnResult(sqlmock.NewResult(0, 3))

	require.NoError(t, repo.MarkAllAsRead(context.Background(), 1))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_CountUnread(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMessageRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM messages WHERE user_id = \\? AND is_read = 0").
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	c, err := repo.CountUnread(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(7), c)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewMessageRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("DELETE FROM messages WHERE id = \\?").
		WithArgs(uint64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 9))
	assert.NoError(t, mock.ExpectationsWereMet())
}
