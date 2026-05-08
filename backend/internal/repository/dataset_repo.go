package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type DatasetRepository struct {
	db *sqlx.DB
}

func NewDatasetRepository(db *sqlx.DB) *DatasetRepository {
	return &DatasetRepository{db: db}
}

func (r *DatasetRepository) Create(ctx context.Context, ds *model.Dataset) error {
	query := `INSERT INTO datasets (name, datasource_id, database_name, table_name, sql, type, mode, status, created_by)
			  VALUES (:name, :datasource_id, :database_name, :table_name, :sql, :type, :mode, :status, :created_by)`
	result, err := r.db.NamedExecContext(ctx, query, ds)
	if err != nil {
		return fmt.Errorf("create dataset failed: %w", err)
	}
	id, _ := result.LastInsertId()
	ds.ID = uint64(id)
	return nil
}

func (r *DatasetRepository) FindByID(ctx context.Context, id uint64) (*model.Dataset, error) {
	var ds model.Dataset
	query := `SELECT * FROM datasets WHERE id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &ds, query, id); err != nil {
		return nil, fmt.Errorf("find dataset failed: %w", err)
	}
	return &ds, nil
}

func (r *DatasetRepository) List(ctx context.Context) ([]model.Dataset, error) {
	var list []model.Dataset
	query := `SELECT * FROM datasets WHERE deleted_at IS NULL ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list datasets failed: %w", err)
	}
	return list, nil
}

func (r *DatasetRepository) Update(ctx context.Context, ds *model.Dataset) error {
	query := `UPDATE datasets SET name = :name, datasource_id = :datasource_id, database_name = :database_name,
			  table_name = :table_name, sql = :sql, type = :type, mode = :mode, status = :status WHERE id = :id`
	if _, err := r.db.NamedExecContext(ctx, query, ds); err != nil {
		return fmt.Errorf("update dataset failed: %w", err)
	}
	return nil
}

func (r *DatasetRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE datasets SET deleted_at = NOW() WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete dataset failed: %w", err)
	}
	return nil
}

func (r *DatasetRepository) CreateFields(ctx context.Context, fields []model.DatasetField) error {
	query := `INSERT INTO dataset_fields (dataset_id, name, type, de_type, length, precision, scale, origin_name)
			  VALUES (:dataset_id, :name, :type, :de_type, :length, :precision, :scale, :origin_name)`
	if _, err := r.db.NamedExecContext(ctx, query, fields); err != nil {
		return fmt.Errorf("create dataset fields failed: %w", err)
	}
	return nil
}

func (r *DatasetRepository) GetFields(ctx context.Context, datasetID uint64) ([]model.DatasetField, error) {
	var fields []model.DatasetField
	query := `SELECT * FROM dataset_fields WHERE dataset_id = ? ORDER BY id`
	if err := r.db.SelectContext(ctx, &fields, query, datasetID); err != nil {
		return nil, fmt.Errorf("get dataset fields failed: %w", err)
	}
	return fields, nil
}

func (r *DatasetRepository) DeleteFields(ctx context.Context, datasetID uint64) error {
	query := `DELETE FROM dataset_fields WHERE dataset_id = ?`
	if _, err := r.db.ExecContext(ctx, query, datasetID); err != nil {
		return fmt.Errorf("delete dataset fields failed: %w", err)
	}
	return nil
}
