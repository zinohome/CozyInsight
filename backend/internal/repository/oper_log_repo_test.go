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

func TestOperationLogRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOperationLogRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO operation_logs").
		WillReturnResult(sqlmock.NewResult(0, 1))

	log := &model.OperationLog{
		UserID: 1, Username: "alice", Method: "POST", Path: "/api/x", IP: "127.0.0.1",
		StatusCode: 200, Duration: 100,
	}
	require.NoError(t, repo.Create(context.Background(), log))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOperationLogRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "user_id", "username", "method", "path", "query", "body", "ip", "user_agent", "status_code", "duration", "error_message", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM operation_logs ORDER BY created_at DESC LIMIT \\?").
		WithArgs(50).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 1, "alice", "GET", "/api/x", "", "", "127.0.0.1", "ua", 200, 50, "", now))

	logs, err := repo.List(context.Background(), 50)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "alice", logs[0].Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogRepository_List_LimitClamped(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOperationLogRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "user_id", "username", "method", "path", "query", "body", "ip", "user_agent", "status_code", "duration", "error_message", "created_at"}
	// limit=0 should be clamped to 100
	mock.ExpectQuery("SELECT \\* FROM operation_logs ORDER BY created_at DESC LIMIT \\?").
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows(columns))

	_, err = repo.List(context.Background(), 0)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogRepository_List_LimitMaxClamped(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOperationLogRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "user_id", "username", "method", "path", "query", "body", "ip", "user_agent", "status_code", "duration", "error_message", "created_at"}
	// limit > 1000 should be clamped to 100
	mock.ExpectQuery("SELECT \\* FROM operation_logs ORDER BY created_at DESC LIMIT \\?").
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows(columns))

	_, err = repo.List(context.Background(), 5000)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
