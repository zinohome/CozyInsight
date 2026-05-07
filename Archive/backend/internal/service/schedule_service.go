package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type ScheduleService interface {
	CreateTask(ctx context.Context, task *model.ScheduleTask) error
	UpdateTask(ctx context.Context, task *model.ScheduleTask) error
	DeleteTask(ctx context.Context, id string) error
	GetTask(ctx context.Context, id string) (*model.ScheduleTask, error)
	ListTasks(ctx context.Context) ([]*model.ScheduleTask, error)
	
	EnableTask(ctx context.Context, id string) error
	DisableTask(ctx context.Context, id string) error
	ExecuteTask(ctx context.Context, id string) error
	
	Start() error
	Stop()
}

type scheduleService struct {
	repo   repository.ScheduleRepository
	cron   *cron.Cron
	jobs   map[string]cron.EntryID
}

func NewScheduleService(repo repository.ScheduleRepository) ScheduleService {
	return &scheduleService{
		repo:   repo,
		cron:   cron.New(),
		jobs:   make(map[string]cron.EntryID),
	}
}

func (s *scheduleService) CreateTask(ctx context.Context, task *model.ScheduleTask) error {
	if task.Name == "" || task.CronExpr == "" {
		return fmt.Errorf("name and cron expression are required")
	}

	// 验证cron表达式
	if _, err := cron.ParseStandard(task.CronExpr); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	task.CreateTime = time.Now().UnixMilli()
	task.UpdateTime = time.Now().UnixMilli()
	task.Status = "inactive"

	if err := s.repo.Create(ctx, task); err != nil {
		return err
	}

	// 如果启用,添加到调度器
	if task.Enabled {
		return s.EnableTask(ctx, task.ID)
	}

	return nil
}

func (s *scheduleService) UpdateTask(ctx context.Context, task *model.ScheduleTask) error {
	existing, err := s.repo.Get(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// 如果cron表达式改变,重新验证
	if task.CronExpr != existing.CronExpr {
		if _, err := cron.ParseStandard(task.CronExpr); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
	}

	task.UpdateTime = time.Now().UnixMilli()
	task.CreateTime = existing.CreateTime

	// 先移除旧任务
	if existing.Enabled {
		s.DisableTask(ctx, task.ID)
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return err
	}

	// 重新添加
	if task.Enabled {
		return s.EnableTask(ctx, task.ID)
	}

	return nil
}

func (s *scheduleService) DeleteTask(ctx context.Context, id string) error {
	s.DisableTask(ctx, id)
	return s.repo.Delete(ctx, id)
}

func (s *scheduleService) GetTask(ctx context.Context, id string) (*model.ScheduleTask, error) {
	return s.repo.Get(ctx, id)
}

func (s *scheduleService) ListTasks(ctx context.Context) ([]*model.ScheduleTask, error) {
	return s.repo.List(ctx)
}

func (s *scheduleService) EnableTask(ctx context.Context, id string) error {
	task, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		s.ExecuteTask(context.Background(), id)
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.jobs[id] = entryID
	task.Enabled = true
	task.Status = "active"
	
	return s.repo.Update(ctx, task)
}

func (s *scheduleService) DisableTask(ctx context.Context, id string) error {
	if entryID, ok := s.jobs[id]; ok {
		s.cron.Remove(entryID)
		delete(s.jobs, id)
	}

	task, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	task.Enabled = false
	task.Status = "inactive"
	
	return s.repo.Update(ctx, task)
}

func (s *scheduleService) ExecuteTask(ctx context.Context, id string) error {
	task, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// 更新最后执行时间
	task.LastRunTime = time.Now().UnixMilli()
	task.Status = "running"
	s.repo.Update(ctx, task)

	// TODO: 根据task.Type执行不同的任务
	// 例如: 发送邮件报告、生成快照等

	task.Status = "active"
	return s.repo.Update(ctx, task)
}

func (s *scheduleService) Start() error {
	// 加载所有启用的任务
	tasks, err := s.repo.List(context.Background())
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.Enabled {
			s.EnableTask(context.Background(), task.ID)
		}
	}

	s.cron.Start()
	return nil
}

func (s *scheduleService) Stop() {
	s.cron.Stop()
}
