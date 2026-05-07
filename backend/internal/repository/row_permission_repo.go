package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type RowPermissionRepository struct {
	db *sqlx.DB
}

func NewRowPermissionRepository(db *sqlx.DB) *RowPermissionRepository {
	return &RowPermissionRepository{db: db}
}

func (r *RowPermissionRepository) Create(ctx context.Context, rp *model.RowPermission) error {
	query := `INSERT INTO row_permissions (dataset_id, field_name, operator, value, user_attr, status)
			  VALUES (:dataset_id, :field_name, :operator, :value, :user_attr, :status)`
	result, err := r.db.NamedExecContext(ctx, query, rp)
	if err != nil {
		return fmt.Errorf("create row permission failed: %w", err)
	}
	id, _ := result.LastInsertId()
	rp.ID = uint64(id)
	return nil
}

func (r *RowPermissionRepository) ListByDataset(ctx context.Context, datasetID uint64) ([]model.RowPermission, error) {
	var list []model.RowPermission
	query := `SELECT * FROM row_permissions WHERE dataset_id = ? AND status = 1`
	if err := r.db.SelectContext(ctx, &list, query, datasetID); err != nil {
		return nil, fmt.Errorf("list row permissions failed: %w", err)
	}
	return list, nil
}

func (r *RowPermissionRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM row_permissions WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete row permission failed: %w", err)
	}
	return nil
}
