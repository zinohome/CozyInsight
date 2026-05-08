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

func TestShareLinkRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewShareLinkRepository(sqlx.NewDb(db, "mysql"))
	mock.ExpectExec("INSERT INTO share_links").
		WillReturnResult(sqlmock.NewResult(1, 1))

	link := &model.ShareLink{
		Token:        "test-token",
		ResourceType: "dashboard",
		ResourceID:   1,
		CreatedBy:    1,
		Status:       1,
	}
	err = repo.Create(context.Background(), link)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), link.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShareLinkRepository_FindByToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewShareLinkRepository(sqlx.NewDb(db, "mysql"))
	columns := []string{"id", "token", "resource_type", "resource_id", "created_by", "expire_at", "status", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM share_links WHERE token = \\? AND status = 1").
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "abc123", "dashboard", 1, 1, nil, 1, now,
		))

	link, err := repo.FindByToken(context.Background(), "abc123")
	require.NoError(t, err)
	assert.Equal(t, "abc123", link.Token)
	assert.Equal(t, uint64(1), link.ResourceID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
