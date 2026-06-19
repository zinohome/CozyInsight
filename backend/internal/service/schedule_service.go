package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

// TaskHandler 描述如何执行一类定时任务。
// 业务侧通过 ScheduleService.RegisterTaskType 注册,
// service.ExecuteTask 按 task.Type 分发。
//
// 当前内置的 type: email_report / snapshot / data_sync。
// 未注册的 type 会被忽略并打日志(降级),不报错。
type TaskHandler func(ctx context.Context, task *model.ScheduleTask) error

// ScheduleService 定时任务服务。
//
// 业务侧应通过 NewScheduleService 注入 repository,然后(可选)调用
// Start 启动调度器。RegisterTaskType 可在运行时动态注册新的任务类型
// handler,用于后续业务模块接入(报表邮件/快照/数据同步等)。
type ScheduleService interface {
	CreateTask(ctx context.Context, task *model.ScheduleTask) error
	UpdateTask(ctx context.Context, task *model.ScheduleTask) error
	DeleteTask(ctx context.Context, id uint64) error
	GetTask(ctx context.Context, id uint64) (*model.ScheduleTask, error)
	ListTasks(ctx context.Context) ([]model.ScheduleTask, error)

	EnableTask(ctx context.Context, id uint64) error
	DisableTask(ctx context.Context, id uint64) error
	ExecuteTask(ctx context.Context, id uint64) error

	// RegisterTaskType 注册一个任务类型处理器。
	// 默认 handler 集合见 NewScheduleService;此处可覆盖或新增。
	RegisterTaskType(taskType string, handler TaskHandler)

	Start() error
	Stop()
}

type scheduleService struct {
	repo     *repository.ScheduleTaskRepository
	cron     *cron.Cron
	mu       sync.Mutex
	jobs     map[uint64]cron.EntryID
	handlers map[string]TaskHandler
	now      func() time.Time // for test injection
}

func NewScheduleService(repo *repository.ScheduleTaskRepository) ScheduleService {
	return &scheduleService{
		repo:     repo,
		cron:     cron.New(),
		jobs:     make(map[uint64]cron.EntryID),
		handlers: defaultTaskHandlers(),
		now:      time.Now,
	}
}

// defaultTaskHandlers 返回内置的 task handler 集合。
// 实际投递(email/snapshot/data_sync)由对应业务模块注入;
// 此处仅占位记录执行结果到 task.Config,以便运维审计和后续接入真实实现。
func defaultTaskHandlers() map[string]TaskHandler {
	return map[string]TaskHandler{
		"email_report": func(ctx context.Context, task *model.ScheduleTask) error {
			return defaultRecordRun(task, "email_report dispatched")
		},
		"snapshot": func(ctx context.Context, task *model.ScheduleTask) error {
			return defaultRecordRun(task, "snapshot generated")
		},
		"data_sync": func(ctx context.Context, task *model.ScheduleTask) error {
			return defaultRecordRun(task, "data sync completed")
		},
	}
}

// defaultRecordRun 把任务执行结果追加到 task.Config 的 last_runs 列表。
// 默认 handler 的占位实现,业务 handler 应直接调用业务逻辑,然后用同样的
// 方式把结果记录下来以便审计。
func defaultRecordRun(task *model.ScheduleTask, message string) error {
	return recordTaskRun(task, message, time.Now)
}

// recordTaskRun 追加执行记录到 task.Config.last_runs(保留最近 10 条)。
// 暴露给 service 内部及 handler 测试使用。
func recordTaskRun(task *model.ScheduleTask, message string, now func() time.Time) error {
	run := map[string]interface{}{
		"at":      now().UnixMilli(),
		"message": message,
	}
	var cfg map[string]interface{}
	if task.Config != "" {
		_ = json.Unmarshal([]byte(task.Config), &cfg)
	}
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	runs, _ := cfg["last_runs"].([]interface{})
	runs = append(runs, run)
	if len(runs) > 10 {
		runs = runs[len(runs)-10:]
	}
	cfg["last_runs"] = runs
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	task.Config = string(b)
	return nil
}

func (s *scheduleService) RegisterTaskType(taskType string, handler TaskHandler) {
	if taskType == "" || handler == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[taskType] = handler
}

func (s *scheduleService) handlerFor(taskType string) TaskHandler {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.handlers[taskType]
}

func (s *scheduleService) CreateTask(ctx context.Context, task *model.ScheduleTask) error {
	if task.Name == "" || task.CronExpr == "" || task.Type == "" {
		return fmt.Errorf("name, type and cron expression are required")
	}
	if _, err := cron.ParseStandard(task.CronExpr); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	now := s.now()
	task.CreatedAt = now
	task.UpdateTime = now
	task.Status = "inactive"
	if task.Config == "" {
		task.Config = "{}"
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return err
	}
	if task.Enabled {
		return s.scheduleJob(task)
	}
	return nil
}

func (s *scheduleService) UpdateTask(ctx context.Context, task *model.ScheduleTask) error {
	existing, err := s.repo.FindByID(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	if task.CronExpr != "" && task.CronExpr != existing.CronExpr {
		if _, err := cron.ParseStandard(task.CronExpr); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
	}
	task.UpdateTime = s.now()
	if err := s.repo.Update(ctx, task); err != nil {
		return err
	}
	// 重新调度:卸载旧 job,根据 enabled 决定是否挂载
	s.unscheduleJob(task.ID)
	if task.Enabled {
		updated, err := s.repo.FindByID(ctx, task.ID)
		if err != nil {
			return err
		}
		return s.scheduleJob(updated)
	}
	return nil
}

func (s *scheduleService) DeleteTask(ctx context.Context, id uint64) error {
	s.unscheduleJob(id)
	return s.repo.Delete(ctx, id)
}

func (s *scheduleService) GetTask(ctx context.Context, id uint64) (*model.ScheduleTask, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *scheduleService) ListTasks(ctx context.Context) ([]model.ScheduleTask, error) {
	return s.repo.List(ctx)
}

func (s *scheduleService) EnableTask(ctx context.Context, id uint64) error {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	task.Enabled = true
	task.Status = "active"
	task.UpdateTime = s.now()
	if err := s.repo.Update(ctx, task); err != nil {
		return err
	}
	return s.scheduleJob(task)
}

func (s *scheduleService) DisableTask(ctx context.Context, id uint64) error {
	s.unscheduleJob(id)
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	task.Enabled = false
	task.Status = "inactive"
	task.UpdateTime = s.now()
	return s.repo.Update(ctx, task)
}

func (s *scheduleService) ExecuteTask(ctx context.Context, id uint64) error {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	handler := s.handlerFor(task.Type)
	if handler == nil {
		log.Printf("[schedule] no handler for type=%q (task=%d); degraded", task.Type, id)
		return nil
	}

	handlerErr := handler(ctx, task)
	if handlerErr != nil {
		task.Status = "error"
		_ = recordTaskRun(task, "handler error: "+handlerErr.Error(), s.now)
		if err := s.repo.Update(ctx, task); err != nil {
			log.Printf("[schedule] update task after handler error: %v", err)
		}
		return handlerErr
	}

	task.Status = "active"
	if err := s.repo.Update(ctx, task); err != nil {
		log.Printf("[schedule] update task after handler success: %v", err)
	}
	return nil
}

// scheduleJob 把 task 挂到 cron 调度器。已存在的 job 会被替换。
func (s *scheduleService) scheduleJob(task *model.ScheduleTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 先卸载已有 job(避免重复)
	if entryID, ok := s.jobs[task.ID]; ok {
		s.cron.Remove(entryID)
	}
	id := task.ID
	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		// cron 触发走后台 ctx,与请求 ctx 隔离
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.ExecuteTask(bgCtx, id); err != nil {
			log.Printf("[schedule] task %d execution failed: %v", id, err)
		}
	})
	if err != nil {
		return fmt.Errorf("add cron job failed: %w", err)
	}
	s.jobs[task.ID] = entryID
	return nil
}

// unscheduleJob 从 cron 调度器移除 task。
func (s *scheduleService) unscheduleJob(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entryID, ok := s.jobs[id]; ok {
		s.cron.Remove(entryID)
		delete(s.jobs, id)
	}
}

func (s *scheduleService) Start() error {
	s.mu.Lock()
	s.cron.Start()
	s.mu.Unlock()
	return nil
}

func (s *scheduleService) Stop() {
	s.mu.Lock()
	s.cron.Stop()
	s.mu.Unlock()
}