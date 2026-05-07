package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
)

type OperLogService interface {
	List(ctx context.Context, userID string, module string, startTime, endTime int64, page, pageSize int) ([]*model.SysOperLog, int64, error)
	CleanOldLogs(ctx context.Context, beforeDays int) error
}

type operLogService struct {
	repo repository.OperLogRepository
}

func NewOperLogService(repo repository.OperLogRepository) OperLogService {
	return &operLogService{repo: repo}
}

func (s *operLogService) List(ctx context.Context, userID string, module string, startTime, endTime int64, page, pageSize int) ([]*model.SysOperLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	
	return s.repo.List(ctx, userID, module, startTime, endTime, page, pageSize)
}

func (s *operLogService) CleanOldLogs(ctx context.Context, beforeDays int) error {
	if beforeDays <= 0 {
		beforeDays = 90 // 默认保留90天
	}
	
	beforeTime := calculateBeforeTime(beforeDays)
	return s.repo.Delete(ctx, beforeTime)
}

func calculateBeforeTime(days int) int64 {
	return (currentTimeMillis() - int64(days)*24*60*60*1000)
}

func currentTimeMillis() int64 {
	return 0 // 实际使用time.Now().UnixMilli()
}
