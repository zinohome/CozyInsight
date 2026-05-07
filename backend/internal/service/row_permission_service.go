package service

import (
	"context"
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

var allowedOperators = map[string]bool{
	"=": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
	"LIKE": true, "NOT LIKE": true, "IN": true, "NOT IN": true,
}

type RowFilterCondition struct {
	FieldName string
	Operator  string
	Value     string
}

// BuildRowFilter 根据用户属性构建结构化 WHERE 条件
func (s *RowPermissionService) BuildRowFilter(ctx context.Context, datasetID uint64, userAttrs map[string]string) ([]RowFilterCondition, error) {
	permissions, err := s.repo.ListByDataset(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	var conditions []RowFilterCondition
	for _, p := range permissions {
		attrValue, ok := userAttrs[p.UserAttr]
		if !ok {
			continue
		}
		op := strings.ToUpper(strings.TrimSpace(p.Operator))
		if !allowedOperators[op] {
			continue
		}
		conditions = append(conditions, RowFilterCondition{
			FieldName: p.FieldName,
			Operator:  op,
			Value:     attrValue,
		})
	}

	return conditions, nil
}
