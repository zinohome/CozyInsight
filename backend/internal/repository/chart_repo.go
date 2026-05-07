package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type ChartRepository struct {
	db *sqlx.DB
}

func NewChartRepository(db *sqlx.DB) *ChartRepository {
	return &ChartRepository{db: db}
}

func (r *ChartRepository) Create(ctx context.Context, chart *model.Chart) error {
	query := `INSERT INTO charts (title, type, dataset_id, config, status, created_by)
			  VALUES (:title, :type, :dataset_id, :config, :status, :created_by)`
	result, err := r.db.NamedExecContext(ctx, query, chart)
	if err != nil {
		return fmt.Errorf("create chart failed: %w", err)
	}
	id, _ := result.LastInsertId()
	chart.ID = uint64(id)
	return nil
}

func (r *ChartRepository) FindByID(ctx context.Context, id uint64) (*model.Chart, error) {
	var chart model.Chart
	query := `SELECT * FROM charts WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &chart, query, id); err != nil {
		return nil, fmt.Errorf("find chart failed: %w", err)
	}
	return &chart, nil
}

func (r *ChartRepository) List(ctx context.Context) ([]model.Chart, error) {
	var list []model.Chart
	query := `SELECT * FROM charts WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list charts failed: %w", err)
	}
	return list, nil
}

func (r *ChartRepository) Update(ctx context.Context, chart *model.Chart) error {
	query := `UPDATE charts SET title = :title, type = :type, dataset_id = :dataset_id, config = :config, status = :status WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, chart); err != nil {
		return fmt.Errorf("update chart failed: %w", err)
	}
	return nil
}

func (r *ChartRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE charts SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete chart failed: %w", err)
	}
	return nil
}
