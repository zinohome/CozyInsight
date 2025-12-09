package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission *model.Permission) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Permission, error)
	List(ctx context.Context) ([]*model.Permission, error)
	
	// 角色权限关联
	GrantPermissionToRole(ctx context.Context, roleID string, permissionIDs []string) error
	RevokePermissionFromRole(ctx context.Context, roleID string, permissionIDs []string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error)
	
	// 资源权限
	GrantResourcePermission(ctx context.Context, rp *model.ResourcePermission) error
	RevokeResourcePermission(ctx context.Context, id string) error
	GetResourcePermissions(ctx context.Context, resourceType, resourceID string) ([]*model.ResourcePermission, error)
	CheckPermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository() PermissionRepository {
	return &permissionRepository{
		db: database.DB,
	}
}

func (r *permissionRepository) Create(ctx context.Context, permission *model.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Permission{}, "id = ?", id).Error
}

func (r *permissionRepository) Get(ctx context.Context, id string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) List(ctx context.Context) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.WithContext(ctx).Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) GrantPermissionToRole(ctx context.Context, roleID string, permissionIDs []string) error {
	var rolePermissions []*model.RolePermission
	for _, permID := range permissionIDs {
		rolePermissions = append(rolePermissions, &model.RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
		})
	}
	return r.db.WithContext(ctx).Create(&rolePermissions).Error
}

func (r *permissionRepository) RevokePermissionFromRole(ctx context.Context, roleID string, permissionIDs []string) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&model.RolePermission{}).Error
}

func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.WithContext(ctx).
		Table("sys_permission").
		Joins("INNER JOIN sys_role_permission ON sys_permission.id = sys_role_permission.permission_id").
		Where("sys_role_permission.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) GrantResourcePermission(ctx context.Context, rp *model.ResourcePermission) error {
	return r.db.WithContext(ctx).Create(rp).Error
}

func (r *permissionRepository) RevokeResourcePermission(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.ResourcePermission{}, "id = ?", id).Error
}

func (r *permissionRepository) GetResourcePermissions(ctx context.Context, resourceType, resourceID string) ([]*model.ResourcePermission, error) {
	var permissions []*model.ResourcePermission
	err := r.db.WithContext(ctx).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) CheckPermission(ctx context.Context, userID, resourceType, resourceID, action string) (bool, error) {
	// 检查用户是否有该资源的权限
	var count int64
	
	// 方式1: 直接授予给用户
	err := r.db.WithContext(ctx).
		Table("sys_resource_permission").
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Where("target_type = ? AND target_id = ?", "user", userID).
		Where("permission = ? OR permission = ?", action, "manage").
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	
	// 方式2: 通过角色授予
	err = r.db.WithContext(ctx).
		Table("sys_resource_permission").
		Joins("INNER JOIN sys_user_role ON sys_resource_permission.target_id = sys_user_role.role_id").
		Where("sys_resource_permission.resource_type = ? AND sys_resource_permission.resource_id = ?", resourceType, resourceID).
		Where("sys_resource_permission.target_type = ?", "role").
		Where("sys_user_role.user_id = ?", userID).
		Where("sys_resource_permission.permission = ? OR sys_resource_permission.permission = ?", action, "manage").
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}
