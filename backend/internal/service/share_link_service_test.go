package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/repository"
)

func TestShareLinkService_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewShareLinkRepository(sqlx.NewDb(db, "mysql"))
	dashboardRepo := repository.NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	svc := NewShareLinkService(repo, dashboardRepo)

	mock.ExpectExec("INSERT INTO share_links").
		WillReturnResult(sqlmock.NewResult(1, 1))

	link, err := svc.Create(context.Background(), "dashboard", 1, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, link.Token)
	assert.Equal(t, "dashboard", link.ResourceType)
	assert.Equal(t, uint64(1), link.ResourceID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShareLinkService_GetDashboard(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewShareLinkRepository(sqlx.NewDb(db, "mysql"))
	dashboardRepo := repository.NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	svc := NewShareLinkService(repo, dashboardRepo)

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

	dashboard, err := svc.GetDashboard(context.Background(), "abc123")
	require.NoError(t, err)
	assert.Equal(t, "Sales Dashboard", dashboard.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShareLinkService_GetDashboard_Expired(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewShareLinkRepository(sqlx.NewDb(db, "mysql"))
	dashboardRepo := repository.NewDashboardRepository(sqlx.NewDb(db, "mysql"))
	svc := NewShareLinkService(repo, dashboardRepo)

	past := time.Now().Add(-time.Hour)
	now := time.Now()
	slColumns := []string{"id", "token", "resource_type", "resource_id", "created_by", "expire_at", "status", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM share_links WHERE token = \\? AND status = 1").
		WithArgs("expired").
		WillReturnRows(sqlmock.NewRows(slColumns).AddRow(
			1, "expired", "dashboard", 1, 1, &past, 1, now,
		))

	_, err = svc.GetDashboard(context.Background(), "expired")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
	assert.NoError(t, mock.ExpectationsWereMet())
}
