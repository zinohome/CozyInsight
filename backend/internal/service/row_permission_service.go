package service

import (
	"context"
	"fmt"
	"strings"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type RowPermissionService struct {
	repo *repository.RowPermissionRepository
}

func NewRowPermissionService(repo *repository.RowPermissionRepository) *RowPermissionService {
	return &RowPermissionService{repo: repo}
}

func (s *RowPermissionService) Create(ctx context.Context, datasetID uint64, fieldName, operator, value, userAttr string) (*model.RowPermission, error) {
	rp := &model.RowPermission{
		DatasetID: datasetID,
		FieldName: fieldName,
		Operator:  operator,
		Value:     value,
		UserAttr:  userAttr,
		Status:    1,
	}
	if err := s.repo.Create(ctx, rp); err != nil {
		return nil, err
	}
	return rp, nil
}

func (s *RowPermissionService) ListByDataset(ctx context.Context, datasetID uint64) ([]model.RowPermission, error) {
	return s.repo.ListByDataset(ctx, datasetID)
}

func (s *RowPermissionService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

// BuildRowFilter 根据用户属性构建 WHERE 条件片段
func (s *RowPermissionService) BuildRowFilter(ctx context.Context, datasetID uint64, userAttrs map[string]string) (string, error) {
	permissions, err := s.repo.ListByDataset(ctx, datasetID)
	if err != nil {
		return "", err
	}

	var conditions []string
	for _, p := range permissions {
		attrValue, ok := userAttrs[p.UserAttr]
		if !ok {
			continue
		}
		cond := fmt.Sprintf("%s %s '%s'", p.FieldName, p.Operator, attrValue)
		conditions = append(conditions, cond)
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return strings.Join(conditions, " AND "), nil
}
