package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type ScheduleTaskRepository struct {
	db *sqlx.DB
}

func NewScheduleTaskRepository(db *sqlx.DB) *ScheduleTaskRepository {
	return &ScheduleTaskRepository{db: db}
}

func (r *ScheduleTaskRepository) Create(ctx context.Context, task *model.ScheduleTask) error {
	query := `INSERT INTO schedule_tasks (name, type, cron_expr, config, enabled, status, created_by, created_at, update_time)
			  VALUES (:name, :type, :cron_expr, :config, :enabled, :status, :created_by, :created_at, :update_time)`
	result, err := r.db.NamedExecContext(ctx, query, task)
	if err != nil {
		return fmt.Errorf("create schedule task failed: %w", err)
	}
	id, _ := result.LastInsertId()
	task.ID = uint64(id)
	return nil
}

func (r *ScheduleTaskRepository) Update(ctx context.Context, task *model.ScheduleTask) error {
	task.UpdateTime = task.UpdateTime // 保留调用方设置的值(由 service 维护)
	if task.UpdateTime.IsZero() {
		// fallback: caller didn't set it
		query := `UPDATE schedule_tasks SET name=:name, type=:type, cron_expr=:cron_expr, config=:config,
		          enabled=:enabled, status=:status, update_time=NOW() WHERE id=:id`
		if _, err := r.db.NamedExecContext(ctx, query, task); err != nil {
			return fmt.Errorf("update schedule task failed: %w", err)
		}
		return nil
	}
	query := `UPDATE schedule_tasks SET name=:name, type=:type, cron_expr=:cron_expr, config=:config,
	          enabled=:enabled, status=:status, update_time=:update_time WHERE id=:id`
	if _, err := r.db.NamedExecContext(ctx, query, task); err != nil {
		return fmt.Errorf("update schedule task failed: %w", err)
	}
	return nil
}

func (r *ScheduleTaskRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM schedule_tasks WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete schedule task failed: %w", err)
	}
	return nil
}

func (r *ScheduleTaskRepository) FindByID(ctx context.Context, id uint64) (*model.ScheduleTask, error) {
	var task model.ScheduleTask
	query := `SELECT * FROM schedule_tasks WHERE id = ?`
	if err := r.db.GetContext(ctx, &task, query, id); err != nil {
		return nil, fmt.Errorf("find schedule task failed: %w", err)
	}
	return &task, nil
}

func (r *ScheduleTaskRepository) List(ctx context.Context) ([]model.ScheduleTask, error) {
	var list []model.ScheduleTask
	query := `SELECT * FROM schedule_tasks ORDER BY id DESC`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list schedule tasks failed: %w", err)
	}
	return list, nil
}

func (r *ScheduleTaskRepository) ListEnabled(ctx context.Context) ([]model.ScheduleTask, error) {
	var list []model.ScheduleTask
	query := `SELECT * FROM schedule_tasks WHERE enabled = 1`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("list enabled schedule tasks failed: %w", err)
	}
	return list, nil
}
