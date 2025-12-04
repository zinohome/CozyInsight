package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

type DatasetFieldService interface {
	CreateField(ctx context.Context, field *model.DatasetTableField) error
	UpdateField(ctx context.Context, field *model.DatasetTableField) error
	DeleteField(ctx context.Context, id string) error
	GetField(ctx context.Context, id string) (*model.DatasetTableField, error)
	ListFields(ctx context.Context, tableId string) ([]*model.DatasetTableField, error)
	SyncFieldsFromTable(ctx context.Context, tableId string) error
}

type datasetFieldService struct {
	repo repository.DatasetFieldRepository
}

func NewDatasetFieldService(repo repository.DatasetFieldRepository) DatasetFieldService {
	return &datasetFieldService{repo: repo}
}

func (s *datasetFieldService) CreateField(ctx context.Context, field *model.DatasetTableField) error {
	if field.ID == "" {
		field.ID = uuid.New().String()
	}
	field.CreateTime = time.Now().UnixMilli()
	field.UpdateTime = time.Now().UnixMilli()

	// 如果没有设置 DisplayName，使用 Name
	if field.DisplayName == "" {
		field.DisplayName = field.Name
	}

	return s.repo.Create(ctx, field)
}

func (s *datasetFieldService) UpdateField(ctx context.Context, field *model.DatasetTableField) error {
	field.UpdateTime = time.Now().UnixMilli()
	return s.repo.Update(ctx, field)
}

func (s *datasetFieldService) DeleteField(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *datasetFieldService) GetField(ctx context.Context, id string) (*model.DatasetTableField, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *datasetFieldService) ListFields(ctx context.Context, tableId string) ([]*model.DatasetTableField, error) {
	return s.repo.ListByTableID(ctx, tableId)
}

// SyncFieldsFromTable 从数据表同步字段（模拟实现）
func (s *datasetFieldService) SyncFieldsFromTable(ctx context.Context, tableId string) error {
	// TODO: 实际应该从 Calcite 或数据库获取真实字段
	// 这里返回模拟字段
	mockFields := []*model.DatasetTableField{
		{
			ID:          uuid.New().String(),
			TableID:     tableId,
			Name:        "id",
			OriginName:  "id",
			Type:        "INTEGER",
			DisplayName: "ID",
			GroupType:   "d",
			SortIndex:   0,
			CreateTime:  time.Now().UnixMilli(),
			UpdateTime:  time.Now().UnixMilli(),
		},
		{
			ID:          uuid.New().String(),
			TableID:     tableId,
			Name:        "name",
			OriginName:  "name",
			Type:        "VARCHAR",
			DisplayName: "名称",
			GroupType:   "d",
			SortIndex:   1,
			CreateTime:  time.Now().UnixMilli(),
			UpdateTime:  time.Now().UnixMilli(),
		},
		{
			ID:          uuid.New().String(),
			TableID:     tableId,
			Name:        "amount",
			OriginName:  "amount",
			Type:        "DECIMAL",
			DisplayName: "金额",
			GroupType:   "q",
			SortIndex:   2,
			CreateTime:  time.Now().UnixMilli(),
			UpdateTime:  time.Now().UnixMilli(),
		},
	}

	return s.repo.BatchCreate(ctx, mockFields)
}
