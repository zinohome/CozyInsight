package service

import (
	"context"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type OperationLogService struct {
	repo *repository.OperationLogRepository
}

func NewOperationLogService(repo *repository.OperationLogRepository) *OperationLogService {
	return &OperationLogService{repo: repo}
}

func (s *OperationLogService) List(ctx context.Context, limit int) ([]model.OperationLog, error) {
	return s.repo.List(ctx, limit)
}
