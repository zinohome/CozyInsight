package repository

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/pkg/database"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Role, error)
	GetByName(ctx context.Context, name string) (*model.Role, error)
	List(ctx context.Context) ([]*model.Role, error)
	
	// 用户角色关联
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error)
	GetRoleUsers(ctx context.Context, roleID string) ([]string, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository() RoleRepository {
	return &roleRepository{
		db: database.DB,
	}
}

func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, "id = ?", id).Error
}

func (r *roleRepository) Get(ctx context.Context, id string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(ctx).Order("create_time DESC").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	userRole := &model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.UserRole{}).Error
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(ctx).
		Table("sys_role").
		Joins("INNER JOIN sys_user_role ON sys_role.id = sys_user_role.role_id").
		Where("sys_user_role.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetRoleUsers(ctx context.Context, roleID string) ([]string, error) {
	var userIDs []string
	err := r.db.WithContext(ctx).
		Table("sys_user_role").
		Where("role_id = ?", roleID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}
