package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RoleService interface {
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Role, error)
	List(ctx context.Context) ([]*model.Role, error)
	
	// 用户角色管理
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error)
	GetRoleUsers(ctx context.Context, roleID string) ([]string, error)
}

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) Create(ctx context.Context, role *model.Role) error {
	if role.Name == "" {
		return fmt.Errorf("role name is required")
	}
	
	// 检查名称是否已存在
	existing, err := s.repo.GetByName(ctx, role.Name)
	if err == nil && existing != nil {
		return fmt.Errorf("role name already exists: %s", role.Name)
	}
	
	if role.ID == "" {
		role.ID = uuid.New().String()
	}
	if role.Type == "" {
		role.Type = "custom"
	}
	role.CreateTime = time.Now().UnixMilli()
	role.UpdateTime = time.Now().UnixMilli()
	
	return s.repo.Create(ctx, role)
}

func (s *roleService) Update(ctx context.Context, role *model.Role) error {
	if role.ID == "" {
		return fmt.Errorf("role id  is required")
	}
	
	// 检查是否存在
	existing, err := s.repo.Get(ctx, role.ID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	
	// 系统角色不允许修改
	if existing.Type == "system" {
		return fmt.Errorf("system role cannot be modified")
	}
	
	role.UpdateTime = time.Now().UnixMilli()
	role.CreateTime = existing.CreateTime // 保留创建时间
	
	return s.repo.Update(ctx, role)
}

func (s *roleService) Delete(ctx context.Context, id string) error {
	// 检查是否存在
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	
	// 系统角色不允许删除
	if existing.Type == "system" {
		return fmt.Errorf("system role cannot be deleted")
	}
	
	return s.repo.Delete(ctx, id)
}

func (s *roleService) Get(ctx context.Context, id string) (*model.Role, error) {
	return s.repo.Get(ctx, id)
}

func (s *roleService) List(ctx context.Context) ([]*model.Role, error) {
	return s.repo.List(ctx)
}

func (s *roleService) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	// 验证角色存在
	if _, err := s.repo.Get(ctx, roleID); err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	
	return s.repo.AssignRoleToUser(ctx, userID, roleID)
}

func (s *roleService) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	return s.repo.RemoveRoleFromUser(ctx, userID, roleID)
}

func (s *roleService) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

func (s *roleService) GetRoleUsers(ctx context.Context, roleID string) ([]string, error) {
	return s.repo.GetRoleUsers(ctx, roleID)
}
