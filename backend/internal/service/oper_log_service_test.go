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

func TestOperationLogService_New(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewOperationLogRepository(db)
	svc := NewOperationLogService(repo)
	assert.NotNil(t, svc)
}

func TestOperationLogService_List(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewOperationLogRepository(db)
	svc := NewOperationLogService(repo)

	columns := []string{"id", "user_id", "username", "method", "path", "query", "body", "ip", "user_agent", "status_code", "duration", "error_message", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM operation_logs ORDER BY created_at DESC LIMIT \\?").
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, 1, "admin", "GET", "/api/v1/user", "", "", "127.0.0.1", "", 200, 10, "", now,
		))

	list, err := svc.List(context.Background(), 0)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "admin", list[0].Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogService_List_WithLimit(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewOperationLogRepository(db)
	svc := NewOperationLogService(repo)

	columns := []string{"id", "user_id", "username", "method", "path", "query", "body", "ip", "user_agent", "status_code", "duration", "error_message", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM operation_logs ORDER BY created_at DESC LIMIT \\?").
		WithArgs(50).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, 1, "admin", "GET", "/api/v1/user", "", "", "127.0.0.1", "", 200, 10, "", now,
		))

	list, err := svc.List(context.Background(), 50)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOperationLogService_List_DBError(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewOperationLogRepository(db)
	svc := NewOperationLogService(repo)

	mock.ExpectQuery("SELECT \\* FROM operation_logs ORDER BY created_at DESC LIMIT \\?").
		WithArgs(100).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.List(context.Background(), 0)
	assert.Error(t, err)
}
