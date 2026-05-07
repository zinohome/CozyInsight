package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type OperationLogRepository struct {
	db *sqlx.DB
}

func NewOperationLogRepository(db *sqlx.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

func (r *OperationLogRepository) Create(ctx context.Context, log *model.OperationLog) error {
	query := `INSERT INTO operation_logs (user_id, username, method, path, query, body, ip, user_agent, status_code, duration, error_message)
			  VALUES (:user_id, :username, :method, :path, :query, :body, :ip, :user_agent, :status_code, :duration, :error_message)`
	_, err := r.db.NamedExecContext(ctx, query, log)
	if err != nil {
		return fmt.Errorf("create operation log failed: %w", err)
	}
	return nil
}

func (r *OperationLogRepository) List(ctx context.Context, limit int) ([]model.OperationLog, error) {
	var list []model.OperationLog
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	query := `SELECT * FROM operation_logs ORDER BY created_at DESC LIMIT ?`
	if err := r.db.SelectContext(ctx, &list, query, limit); err != nil {
		return nil, fmt.Errorf("list operation logs failed: %w", err)
	}
	return list, nil
}
