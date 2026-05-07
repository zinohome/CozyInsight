package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
)

type DatasetGroupService interface {
	CreateGroup(ctx context.Context, group *model.DatasetGroup) error
	UpdateGroup(ctx context.Context, group *model.DatasetGroup) error
	DeleteGroup(ctx context.Context, id string) error
	GetGroup(ctx context.Context, id string) (*model.DatasetGroup, error)
	ListGroups(ctx context.Context) ([]*model.DatasetGroup, error)

	// 获取树形结构
	GetGroupTree(ctx context.Context) ([]*model.DatasetGroupTree, error)
}

type datasetGroupService struct {
	repo repository.DatasetRepository
}

func NewDatasetGroupService(repo repository.DatasetRepository) DatasetGroupService {
	return &datasetGroupService{repo: repo}
}

func (s *datasetGroupService) CreateGroup(ctx context.Context, group *model.DatasetGroup) error {
	return s.repo.CreateGroup(ctx, group)
}

func (s *datasetGroupService) UpdateGroup(ctx context.Context, group *model.DatasetGroup) error {
	return s.repo.UpdateGroup(ctx, group)
}

func (s *datasetGroupService) DeleteGroup(ctx context.Context, id string) error {
	return s.repo.DeleteGroup(ctx, id)
}

func (s *datasetGroupService) GetGroup(ctx context.Context, id string) (*model.DatasetGroup, error) {
	return s.repo.GetGroup(ctx, id)
}

func (s *datasetGroupService) ListGroups(ctx context.Context) ([]*model.DatasetGroup, error) {
	return s.repo.ListGroups(ctx)
}

func (s *datasetGroupService) GetGroupTree(ctx context.Context) ([]*model.DatasetGroupTree, error) {
	groups, err := s.repo.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	return buildTree(groups, "0"), nil
}

// buildTree 构建树形结构
func buildTree(groups []*model.DatasetGroup, pid string) []*model.DatasetGroupTree {
	var result []*model.DatasetGroupTree

	for _, group := range groups {
		if group.PID == pid {
			node := &model.DatasetGroupTree{
				ID:    group.ID,
				Name:  group.Name,
				Pid:   group.PID,
				Level: group.Level,
				Type:  group.Type,
			}

			children := buildTree(groups, group.ID)
			if len(children) > 0 {
				node.Children = children
			}

			result = append(result, node)
		}
	}

	return result
}
