package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/testutil"
)

func TestDashboardService_Create(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	mock.ExpectExec("INSERT INTO dashboards").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateDashboardRequest{
		Title:  "Test Dashboard",
		Config: "{}",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "Test Dashboard", result.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Create_DefaultType(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	mock.ExpectExec("INSERT INTO dashboards").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateDashboardRequest{
		Title:  "Test Dashboard",
		Config: "{}",
		Type:   "",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, model.DashboardTypeDashboard, result.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Create_ScreenType(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	mock.ExpectExec("INSERT INTO dashboards").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := svc.Create(context.Background(), &dto.CreateDashboardRequest{
		Title:  "Test Screen",
		Config: "{}",
		Type:   model.DashboardTypeScreen,
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, model.DashboardTypeScreen, result.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Create_InvalidType(t *testing.T) {
	db, _ := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	_, err := svc.Create(context.Background(), &dto.CreateDashboardRequest{
		Title:  "Test",
		Config: "{}",
		Type:   "invalid",
	}, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid dashboard type")
}

func TestDashboardService_GetByID(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	columns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test Dashboard", "dashboard", "{}", "", 0, 1, 1, now, now, nil,
		))

	d, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), d.ID)
	assert.Equal(t, "Test Dashboard", d.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_List(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	columns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "Dashboard1", "dashboard", "{}", "", 0, 1, 1, now, now, nil).
			AddRow(2, "Dashboard2", "dashboard", "{}", "", 0, 1, 1, now, now, nil))

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Update(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	columns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old", "dashboard", "{}", "", 0, 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE dashboards SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	status := int8(0)
	err := svc.Update(context.Background(), 1, &dto.UpdateDashboardRequest{
		Title:  "Updated",
		Config: `{"color":"blue"}`,
		Status: &status,
	}, 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Delete(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	columns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test", "dashboard", "{}", "", 0, 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE dashboards SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_AddChart(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	mock.ExpectExec("INSERT INTO dashboard_charts").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.AddChart(context.Background(), 1, &dto.AddChartToDashboardRequest{
		ChartID:   1,
		PositionX: 0,
		PositionY: 0,
		Width:     6,
		Height:    4,
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_GetCharts(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	columns := []string{"id", "dashboard_id", "chart_id", "position_x", "position_y", "width", "height", "config", "created_at", "updated_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboard_charts WHERE dashboard_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, 1, 1, 0, 0, 6, 4, "{}", now, now,
		))

	list, err := svc.GetCharts(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, uint64(1), list[0].ChartID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_Update_Forbidden(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	columns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old", "dashboard", "{}", "", 0, 1, 2, now, now, nil,
		))

	err := svc.Update(context.Background(), 1, &dto.UpdateDashboardRequest{
		Title: "Updated",
	}, 1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotOwner))
}

func TestDashboardService_Delete_Forbidden(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	columns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old", "dashboard", "{}", "", 0, 1, 2, now, now, nil,
		))

	err := svc.Delete(context.Background(), 1, 1)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotOwner))
}

func TestDashboardService_EnableShare(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	shareLinkRepo := repository.NewShareLinkRepository(db)
	svc := NewDashboardService(repo, shareLinkRepo)

	dbColumns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dbColumns).AddRow(
			1, "Test", "dashboard", "{}", "", 0, 1, 1, now, now, nil,
		))

	slColumns := []string{"id", "token", "resource_type", "resource_id", "created_by", "expire_at", "status", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM share_links WHERE resource_type = \\? AND resource_id = \\? AND status = 1").
		WithArgs("dashboard", 1).
		WillReturnRows(sqlmock.NewRows(slColumns))

	mock.ExpectExec("INSERT INTO share_links").
		WillReturnResult(sqlmock.NewResult(1, 1))

	token, err := svc.EnableShare(context.Background(), 1, 1, "", 0)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_DisableShare(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	shareLinkRepo := repository.NewShareLinkRepository(db)
	svc := NewDashboardService(repo, shareLinkRepo)

	dbColumns := []string{"id", "title", "type", "config", "share_token", "share_enabled", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM dashboards WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dbColumns).AddRow(
			1, "Test", "dashboard", "{}", "", 0, 1, 1, now, now, nil,
		))

	slColumns := []string{"id", "token", "resource_type", "resource_id", "created_by", "expire_at", "status", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM share_links WHERE resource_type = \\? AND resource_id = \\? AND status = 1").
		WithArgs("dashboard", 1).
		WillReturnRows(sqlmock.NewRows(slColumns).AddRow(
			1, "abc123", "dashboard", 1, 1, nil, 1, now,
		))

	mock.ExpectExec("UPDATE share_links SET status = 0 WHERE id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.DisableShare(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardService_RemoveChart(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDashboardRepository(db)
	svc := NewDashboardService(repo, nil)

	mock.ExpectExec("DELETE FROM dashboard_charts WHERE dashboard_id = \\? AND chart_id = \\?").
		WithArgs(1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.RemoveChart(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
