package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type DatasourceRepository struct {
	db *sqlx.DB
}

func NewDatasourceRepository(db *sqlx.DB) *DatasourceRepository {
	return &DatasourceRepository{db: db}
}

func (r *DatasourceRepository) Create(ctx context.Context, ds *model.Datasource) error {
	query := `INSERT INTO datasources (name, type, config, file_path, file_type, status, created_by)
			  VALUES (:name, :type, :config, :file_path, :file_type, :status, :created_by)`
	result, err := r.db.NamedExecContext(ctx, query, ds)
	if err != nil {
		return fmt.Errorf("create datasource failed: %w", err)
	}
	id, _ := result.LastInsertId()
	ds.ID = uint64(id)
	return nil
}

func (r *DatasourceRepository) FindByID(ctx context.Context, id uint64) (*model.Datasource, error) {
	var ds model.Datasource
	query := `SELECT * FROM datasources WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &ds, query, id); err != nil {
		return nil, fmt.Errorf("find datasource failed: %w", err)
	}
	return &ds, nil
}

func (r *DatasourceRepository) List(ctx context.Context) ([]model.Datasource, error) {
	var list []model.Datasource
	query := `SELECT * FROM datasources WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list datasources failed: %w", err)
	}
	return list, nil
}

func (r *DatasourceRepository) Update(ctx context.Context, ds *model.Datasource) error {
	query := `UPDATE datasources SET name = :name, type = :type, config = :config, file_path = :file_path, file_type = :file_type, status = :status WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, ds); err != nil {
		return fmt.Errorf("update datasource failed: %w", err)
	}
	return nil
}

func (r *DatasourceRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE datasources SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete datasource failed: %w", err)
	}
	return nil
}
