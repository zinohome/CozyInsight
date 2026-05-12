package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/dto"
)

type WorkbenchRepository struct {
	db *sqlx.DB
}

func NewWorkbenchRepository(db *sqlx.DB) *WorkbenchRepository {
	return &WorkbenchRepository{db: db}
}

func (r *WorkbenchRepository) CountByCreatedBy(ctx context.Context, table string, userID uint64, extraWhere string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE created_by = ? AND deleted_at IS NULL", table)
	if extraWhere != "" {
		query += " AND " + extraWhere
	}
	var count int64
	if err := r.db.GetContext(ctx, &count, query, userID); err != nil {
		return 0, fmt.Errorf("count %s failed: %w", table, err)
	}
	return count, nil
}

func (r *WorkbenchRepository) UpsertRecentView(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	query := `INSERT INTO recent_views (user_id, resource_type, resource_id, visited_at)
			  VALUES (?, ?, ?, NOW())
			  ON DUPLICATE KEY UPDATE visited_at = NOW()`
	if _, err := r.db.ExecContext(ctx, query, userID, resourceType, resourceID); err != nil {
		return fmt.Errorf("upsert recent view failed: %w", err)
	}
	return nil
}

func (r *WorkbenchRepository) ListRecentViews(ctx context.Context, userID uint64, limit int) ([]dto.RecentViewItem, error) {
	query := `SELECT d.id, d.title, d.type, rv.visited_at
			  FROM recent_views rv
			  JOIN dashboards d ON rv.resource_id = d.id AND d.deleted_at IS NULL
			  WHERE rv.user_id = ? AND rv.resource_type = d.type
			  ORDER BY rv.visited_at DESC
			  LIMIT ?`
	var list []dto.RecentViewItem
	if err := r.db.SelectContext(ctx, &list, query, userID, limit); err != nil {
		return nil, fmt.Errorf("list recent views failed: %w", err)
	}
	return list, nil
}

func (r *WorkbenchRepository) AddFavorite(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	query := `INSERT INTO favorites (user_id, resource_type, resource_id) VALUES (?, ?, ?)
			  ON DUPLICATE KEY UPDATE created_at = NOW()`
	if _, err := r.db.ExecContext(ctx, query, userID, resourceType, resourceID); err != nil {
		return fmt.Errorf("add favorite failed: %w", err)
	}
	return nil
}

func (r *WorkbenchRepository) DeleteFavorite(ctx context.Context, userID uint64, resourceType string, resourceID uint64) error {
	query := `DELETE FROM favorites WHERE user_id = ? AND resource_type = ? AND resource_id = ?`
	if _, err := r.db.ExecContext(ctx, query, userID, resourceType, resourceID); err != nil {
		return fmt.Errorf("delete favorite failed: %w", err)
	}
	return nil
}

func (r *WorkbenchRepository) ListFavorites(ctx context.Context, userID uint64) ([]dto.FavoriteItem, error) {
	query := `SELECT d.id, d.title, d.type, f.created_at
			  FROM favorites f
			  JOIN dashboards d ON f.resource_id = d.id AND d.deleted_at IS NULL
			  WHERE f.user_id = ? AND f.resource_type IN ('dashboard', 'screen') AND d.type = f.resource_type
			  ORDER BY f.created_at DESC`
	var list []dto.FavoriteItem
	if err := r.db.SelectContext(ctx, &list, query, userID); err != nil {
		return nil, fmt.Errorf("list favorites failed: %w", err)
	}
	return list, nil
}
