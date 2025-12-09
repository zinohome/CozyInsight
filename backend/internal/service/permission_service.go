package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PermissionService interface {
	Create(ctx context.Context, permission *model.Permission) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Permission, error)
	List(ctx context.Context) ([]*model.Permission, error)
	
	// 角色权限管理
	GrantPermissionToRole(ctx context.Context, roleID string, permissionIDs []string) error
	RevokePermissionFromRole(ctx context.Context, roleID string, permissionIDs []string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error)
	
	// 资源权限管理
	GrantResourcePermission(ctx context.Context, resourceType, resourceID, targetType, targetID, permission, createBy string) error
	RevokeResourcePermission(ctx context.Context, id string) error
	GetResourcePermissions(ctx context.Context, resourceType, resourceID string) ([]*model.ResourcePermission, error)
	
	// 权限检查
	CheckPermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error)
	CheckUserHasRole(ctx context.Context, userID, roleName string) (bool, error)
}

type permissionService struct {
	repo     repository.PermissionRepository
	roleRepo repository.RoleRepository
}

func NewPermissionService(repo repository.PermissionRepository, roleRepo repository.RoleRepository) PermissionService {
	return &permissionService{
		repo:     repo,
		roleRepo: roleRepo,
	}
}

func (s *permissionService) Create(ctx context.Context, permission *model.Permission) error {
	if permission.Name == "" {
		return fmt.Errorf("permission name is required")
	}
	
	if permission.ID == "" {
		permission.ID = uuid.New().String()
	}
	permission.CreateTime = time.Now().UnixMilli()
	
	return s.repo.Create(ctx, permission)
}

func (s *permissionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *permissionService) Get(ctx context.Context, id string) (*model.Permission, error) {
	return s.repo.Get(ctx, id)
}

func (s *permissionService) List(ctx context.Context) ([]*model.Permission, error) {
	return s.repo.List(ctx)
}

func (s *permissionService) GrantPermissionToRole(ctx context.Context, roleID string, permissionIDs []string) error {
	// 验证角色存在
	if _, err := s.roleRepo.Get(ctx, roleID); err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	
	// 验证所有权限存在
	for _, permID := range permissionIDs {
		if _, err := s.repo.Get(ctx, permID); err != nil {
			return fmt.Errorf("permission not found: %s", permID)
		}
	}
	
	return s.repo.GrantPermissionToRole(ctx, roleID, permissionIDs)
}

func (s *permissionService) RevokePermissionFromRole(ctx context.Context, roleID string, permissionIDs []string) error {
	return s.repo.RevokePermissionFromRole(ctx, roleID, permissionIDs)
}

func (s *permissionService) GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error) {
	return s.repo.GetRolePermissions(ctx, roleID)
}

func (s *permissionService) GrantResourcePermission(ctx context.Context, resourceType, resourceID, targetType, targetID, permission, createBy string) error {
	// 验证参数
	if resourceType == "" || resourceID == "" || targetType == "" || targetID == "" || permission == "" {
		return fmt.Errorf("all parameters are required")
	}
	
	validTargetTypes := map[string]bool{"user": true, "role": true}
	if !validTargetTypes[targetType] {
		return fmt.Errorf("invalid target type: %s", targetType)
	}
	
	validPermissions := map[string]bool{"read": true, "write": true, "manage": true}
	if !validPermissions[permission] {
		return fmt.Errorf("invalid permission: %s", permission)
	}
	
	rp := &model.ResourcePermission{
		ID:           uuid.New().String(),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TargetType:   targetType,
		TargetID:     targetID,
		Permission:   permission,
		CreateTime:   time.Now().UnixMilli(),
		CreateBy:     createBy,
	}
	
	return s.repo.GrantResourcePermission(ctx, rp)
}

func (s *permissionService) RevokeResourcePermission(ctx context.Context, id string) error {
	return s.repo.RevokeResourcePermission(ctx, id)
}

func (s *permissionService) GetResourcePermissions(ctx context.Context, resourceType, resourceID string) ([]*model.ResourcePermission, error) {
	return s.repo.GetResourcePermissions(ctx, resourceType, resourceID)
}

func (s *permissionService) CheckPermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error) {
	return s.repo.CheckPermission(ctx, userID, resourceType, resourceID, action)
}

func (s *permissionService) CheckUserHasRole(ctx context.Context, userID, roleName string) (bool, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}
	
	for _, role := range roles {
		if role.Name == roleName {
			return true, nil
		}
	}
	
	return false, nil
}
