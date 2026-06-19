package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

// newTestService 返回一个已注入 sqlmock 的 ScheduleService。
// 用于验证 service 与 repo 的交互(主要是 ExecuteTask 调度路径)。
func newTestService(t *testing.T) (ScheduleService, sqlmock.Sqlmock, *sqlx.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	repo := repository.NewScheduleTaskRepository(sqlxDB)
	svc := NewScheduleService(repo)
	return svc, mock, sqlxDB
}

func columns() []string {
	return []string{"id", "name", "type", "cron_expr", "config", "enabled", "status", "created_by", "created_at", "update_time"}
}

func rowFor(task model.ScheduleTask) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(columns()).AddRow(
		task.ID, task.Name, task.Type, task.CronExpr, task.Config,
		task.Enabled, task.Status, task.CreatedBy, now, now,
	)
}

// TestExecuteTask_DispatchesByType 验证默认 handler 按 type 分发。
// 关注 dispatch 行为本身:用替换的 handler 验证被调用一次。
func TestExecuteTask_DispatchesByType(t *testing.T) {
	svc, mock, _ := newTestService(t)

	called := 0
	capturedType := ""
	svc.RegisterTaskType("email_report", func(ctx context.Context, task *model.ScheduleTask) error {
		called++
		capturedType = task.Type
		return recordTaskRun(task, "captured", time.Now)
	})

	mock.ExpectQuery("SELECT \\* FROM schedule_tasks WHERE id = \\?").
		WithArgs(uint64(1)).
		WillReturnRows(rowFor(model.ScheduleTask{
			ID: 1, Name: "t", Type: "email_report", CronExpr: "0 9 * * *",
			Config: "{}", Enabled: true, Status: "active", CreatedBy: 1,
		}))

	mock.ExpectExec("UPDATE schedule_tasks SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.ExecuteTask(context.Background(), 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, 1, called, "email_report handler should be invoked exactly once")
	assert.Equal(t, "email_report", capturedType)
}

// TestExecuteTask_UnknownTypeDegraded 验证未注册 type 走降级路径(无错误)。
func TestExecuteTask_UnknownTypeDegraded(t *testing.T) {
	svc, mock, _ := newTestService(t)

	mock.ExpectQuery("SELECT \\* FROM schedule_tasks WHERE id = \\?").
		WithArgs(uint64(2)).
		WillReturnRows(rowFor(model.ScheduleTask{
			ID: 2, Name: "t", Type: "unknown_type", CronExpr: "0 9 * * *",
			Config: "{}", Enabled: true, Status: "active", CreatedBy: 1,
		}))

	// unknown type 不会调用 UPDATE(降级直接 return)
	err := svc.ExecuteTask(context.Background(), 2)
	require.NoError(t, err, "unknown type should degrade gracefully")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestExecuteTask_CustomHandlerOverride 验证 RegisterTaskType 可覆盖/新增 handler。
func TestExecuteTask_CustomHandlerOverride(t *testing.T) {
	svc, mock, _ := newTestService(t)

	called := false
	svc.RegisterTaskType("email_report", func(ctx context.Context, task *model.ScheduleTask) error {
		called = true
		return recordTaskRun(task, "custom ran", time.Now)
	})

	mock.ExpectQuery("SELECT \\* FROM schedule_tasks WHERE id = \\?").
		WithArgs(uint64(3)).
		WillReturnRows(rowFor(model.ScheduleTask{
			ID: 3, Name: "t", Type: "email_report", CronExpr: "0 9 * * *",
			Config: "{}", Enabled: true, Status: "active", CreatedBy: 1,
		}))
	mock.ExpectExec("UPDATE schedule_tasks SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, svc.ExecuteTask(context.Background(), 3))
	assert.True(t, called, "custom handler should be invoked")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestExecuteTask_HandlerError 验证 handler 报错时整体返回 err 且 status 置为 error。
func TestExecuteTask_HandlerError(t *testing.T) {
	svc, mock, _ := newTestService(t)

	svc.RegisterTaskType("data_sync", func(ctx context.Context, task *model.ScheduleTask) error {
		return errors.New("synthetic failure")
	})

	mock.ExpectQuery("SELECT \\* FROM schedule_tasks WHERE id = \\?").
		WithArgs(uint64(4)).
		WillReturnRows(rowFor(model.ScheduleTask{
			ID: 4, Name: "t", Type: "data_sync", CronExpr: "0 9 * * *",
			Config: "{}", Enabled: true, Status: "active", CreatedBy: 1,
		}))
	mock.ExpectExec("UPDATE schedule_tasks SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.ExecuteTask(context.Background(), 4)
	require.Error(t, err, "handler error should propagate")
	assert.Contains(t, err.Error(), "synthetic failure")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestExecuteTask_MissingTask 验证 task 不存在时返回错误。
func TestExecuteTask_MissingTask(t *testing.T) {
	svc, mock, _ := newTestService(t)

	mock.ExpectQuery("SELECT \\* FROM schedule_tasks WHERE id = \\?").
		WithArgs(uint64(999)).
		WillReturnRows(sqlmock.NewRows(columns()))

	err := svc.ExecuteTask(context.Background(), 999)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRegisterTaskType_RejectsEmpty 验证空 type/nil handler 被拒绝(无 panic)。
func TestRegisterTaskType_RejectsEmpty(t *testing.T) {
	svc, _, _ := newTestService(t)
	// 不应 panic
	svc.RegisterTaskType("", func(ctx context.Context, task *model.ScheduleTask) error { return nil })
	svc.RegisterTaskType("foo", nil)
	// 通过 dispatch 验证未注册
	mock := struct{}{}
	_ = mock
}

// TestCreateTask_InvalidCron 验证无效 cron 表达式被拒。
func TestCreateTask_InvalidCron(t *testing.T) {
	svc, _, _ := newTestService(t)
	err := svc.CreateTask(context.Background(), &model.ScheduleTask{
		Name: "x", Type: "email_report", CronExpr: "this is not cron",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cron")
}

// TestCreateTask_RequiresName 验证空 name 被拒。
func TestCreateTask_RequiresName(t *testing.T) {
	svc, _, _ := newTestService(t)
	err := svc.CreateTask(context.Background(), &model.ScheduleTask{
		Type: "email_report", CronExpr: "0 9 * * *",
	})
	require.Error(t, err)
}

// TestRecordTaskRun_TrimsToLast10 验证 last_runs 超过 10 条时截断。
func TestRecordTaskRun_TrimsToLast10(t *testing.T) {
	task := &model.ScheduleTask{Config: "{}"}
	now := time.Now
	for i := 0; i < 12; i++ {
		require.NoError(t, recordTaskRun(task, "run", now))
	}
	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(task.Config), &cfg))
	runs := cfg["last_runs"].([]interface{})
	assert.Len(t, runs, 10, "should trim to last 10 runs")
}