package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type DashboardRepository struct {
	db *sqlx.DB
}

func NewDashboardRepository(db *sqlx.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) Create(ctx context.Context, d *model.Dashboard) error {
	query := `INSERT INTO dashboards (title, config, share_token, share_enabled, status, created_by)
			  VALUES (:title, :config, :share_token, :share_enabled, :status, :created_by)`
	result, err := r.db.NamedExecContext(ctx, query, d)
	if err != nil {
		return fmt.Errorf("create dashboard failed: %w", err)
	}
	id, _ := result.LastInsertId()
	d.ID = uint64(id)
	return nil
}

func (r *DashboardRepository) FindByID(ctx context.Context, id uint64) (*model.Dashboard, error) {
	var d model.Dashboard
	query := `SELECT * FROM dashboards WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &d, query, id); err != nil {
		return nil, fmt.Errorf("find dashboard failed: %w", err)
	}
	return &d, nil
}

func (r *DashboardRepository) List(ctx context.Context) ([]model.Dashboard, error) {
	var list []model.Dashboard
	query := `SELECT * FROM dashboards WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list dashboards failed: %w", err)
	}
	return list, nil
}

func (r *DashboardRepository) Update(ctx context.Context, d *model.Dashboard) error {
	query := `UPDATE dashboards SET title = :title, config = :config, share_token = :share_token, share_enabled = :share_enabled, status = :status WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, d); err != nil {
		return fmt.Errorf("update dashboard failed: %w", err)
	}
	return nil
}

func (r *DashboardRepository) FindByShareToken(ctx context.Context, token string) (*model.Dashboard, error) {
	var d model.Dashboard
	query := `SELECT * FROM dashboards WHERE share_token = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &d, query, token); err != nil {
		return nil, fmt.Errorf("find dashboard by share token failed: %w", err)
	}
	return &d, nil
}

func (r *DashboardRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE dashboards SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete dashboard failed: %w", err)
	}
	return nil
}

func (r *DashboardRepository) AddChart(ctx context.Context, dc *model.DashboardChart) error {
	query := `INSERT INTO dashboard_charts (dashboard_id, chart_id, position_x, position_y, width, height, config)
			  VALUES (:dashboard_id, :chart_id, :position_x, :position_y, :width, :height, :config)`
	_, err := r.db.NamedExecContext(ctx, query, dc)
	if err != nil {
		return fmt.Errorf("add chart to dashboard failed: %w", err)
	}
	return nil
}

func (r *DashboardRepository) GetCharts(ctx context.Context, dashboardID uint64) ([]model.DashboardChart, error) {
	var list []model.DashboardChart
	query := `SELECT * FROM dashboard_charts WHERE dashboard_id = ? ORDER BY id`
	if err := r.db.SelectContext(ctx, &list, query, dashboardID); err != nil {
		return nil, fmt.Errorf("get dashboard charts failed: %w", err)
	}
	return list, nil
}

func (r *DashboardRepository) RemoveChart(ctx context.Context, dashboardID, chartID uint64) error {
	query := `DELETE FROM dashboard_charts WHERE dashboard_id = ? AND chart_id = ?`
	if _, err := r.db.ExecContext(ctx, query, dashboardID, chartID); err != nil {
		return fmt.Errorf("remove chart from dashboard failed: %w", err)
	}
	return nil
}
