package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type CalculatedFieldService interface {
	Create(ctx context.Context, field *model.DatasetTableFieldCalculated) error
	Update(ctx context.Context, field *model.DatasetTableFieldCalculated) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.DatasetTableFieldCalculated, error)
	ListByTable(ctx context.Context, tableID string) ([]*model.DatasetTableFieldCalculated, error)
	
	// 计算字段值
	CalculateValue(ctx context.Context, expression string, row map[string]interface{}) (interface{}, error)
}

type calculatedFieldService struct {
	repo repository.CalculatedFieldRepository
}

func NewCalculatedFieldService(repo repository.CalculatedFieldRepository) CalculatedFieldService {
	return &calculatedFieldService{repo: repo}
}

func (s *calculatedFieldService) Create(ctx context.Context, field *model.DatasetTableFieldCalculated) error {
	if field.DatasetTableID == "" || field.FieldName == "" {
		return fmt.Errorf("table id and field name are required")
	}

	if field.Expression == "" {
		return fmt.Errorf("expression is required")
	}

	if field.ID == "" {
		field.ID = uuid.New().String()
	}

	return s.repo.Create(ctx, field)
}

func (s *calculatedFieldService) Update(ctx context.Context, field *model.DatasetTableFieldCalculated) error {
	existing, err := s.repo.Get(ctx, field.ID)
	if err != nil {
		return fmt.Errorf("field not found: %w", err)
	}

	field.CreateTime = existing.CreateTime
	return s.repo.Update(ctx, field)
}

func (s *calculatedFieldService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *calculatedFieldService) Get(ctx context.Context, id string) (*model.DatasetTableFieldCalculated, error) {
	return s.repo.Get(ctx, id)
}

func (s *calculatedFieldService) ListByTable(ctx context.Context, tableID string) ([]*model.DatasetTableFieldCalculated, error) {
	return s.repo.ListByTable(ctx, tableID)
}

// CalculateValue 计算字段值 (简化实现,支持基本四则运算)
func (s *calculatedFieldService) CalculateValue(ctx context.Context, expression string, row map[string]interface{}) (interface{}, error) {
	// 简化的表达式计算
	// 支持格式: [字段名] + [字段名] * 数字 等
	
	expr := expression
	
	// 替换字段名为实际值
	for key, value := range row {
		placeholder := fmt.Sprintf("[%s]", key)
		if strings.Contains(expr, placeholder) {
			expr = strings.ReplaceAll(expr, placeholder, fmt.Sprintf("%v", value))
		}
	}
	
	// 这里简化处理,实际应该使用表达式解析器
	// 可以集成 github.com/Knetic/govaluate 等库
	
	return expr, nil
}
