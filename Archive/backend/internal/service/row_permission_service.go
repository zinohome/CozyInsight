package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RowPermissionService interface {
	Create(ctx context.Context, permission *model.DatasetRowPermissions) error
	Update(ctx context.Context, permission *model.DatasetRowPermissions) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.DatasetRowPermissions, error)
	ListByDataset(ctx context.Context, datasetID string) ([]*model.DatasetRowPermissions, error)
	
	// 获取用户对数据集的行权限WHERE条件
	GetUserRowPermissionWhere(ctx context.Context, userID, datasetID string) (string, error)
}

type rowPermissionService struct {
	repo     repository.RowPermissionRepository
	roleRepo repository.RoleRepository
}

func NewRowPermissionService(
	repo repository.RowPermissionRepository,
	roleRepo repository.RoleRepository,
) RowPermissionService {
	return &rowPermissionService{
		repo:     repo,
		roleRepo: roleRepo,
	}
}

func (s *rowPermissionService) Create(ctx context.Context, permission *model.DatasetRowPermissions) error {
	if permission.DatasetID == "" || permission.AuthTargetID == "" {
		return fmt.Errorf("datasetId and authTargetId are required")
	}

	if permission.ID == "" {
		permission.ID = uuid.New().String()
	}
	permission.CreateTime = time.Now().UnixMilli()
	permission.UpdateTime = time.Now().UnixMilli()

	return s.repo.Create(ctx, permission)
}

func (s *rowPermissionService) Update(ctx context.Context, permission *model.DatasetRowPermissions) error {
	existing, err := s.repo.Get(ctx, permission.ID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	permission.CreateTime = existing.CreateTime
	permission.UpdateTime = time.Now().UnixMilli()

	return s.repo.Update(ctx, permission)
}

func (s *rowPermissionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *rowPermissionService) Get(ctx context.Context, id string) (*model.DatasetRowPermissions, error) {
	return s.repo.Get(ctx, id)
}

func (s *rowPermissionService) ListByDataset(ctx context.Context, datasetID string) ([]*model.DatasetRowPermissions, error) {
	return s.repo.ListByDataset(ctx, datasetID)
}

func (s *rowPermissionService) GetUserRowPermissionWhere(ctx context.Context, userID, datasetID string) (string, error) {
	// 获取所有该数据集的行权限
	permissions, err := s.repo.ListByDataset(ctx, datasetID)
	if err != nil {
		return "", err
	}

	var whereClauses []string

	for _, perm := range permissions {
		if !perm.Enable {
			continue
		}

		// 检查是否适用于该用户
		applies := false
		
		switch perm.AuthTargetType {
		case "user":
			if perm.AuthTargetID == userID {
				applies = true
			}
		case "role":
			// 检查用户是否有该角色
			roles, err := s.roleRepo.GetUserRoles(ctx, userID)
			if err == nil {
				for _, role := range roles {
					if role.ID == perm.AuthTargetID {
						applies = true
						break
					}
				}
			}
		}

		if applies && perm.WhereCondition != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", perm.WhereCondition))
		}
	}

	// 如果没有行权限,返回空
	if len(whereClauses) == 0 {
		return "", nil
	}

	// 多个行权限用OR连接
	return strings.Join(whereClauses, " OR "), nil
}
