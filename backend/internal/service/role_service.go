package service

import (
	"context"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type RoleService struct {
	repo *repository.RoleRepository
}

func NewRoleService(repo *repository.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) Create(ctx context.Context, req *dto.CreateRoleRequest) (*model.Role, error) {
	role := &model.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
	}
	if role.Status == 0 {
		role.Status = 1
	}
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RoleService) GetByID(ctx context.Context, id uint64) (*model.Role, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *RoleService) List(ctx context.Context) ([]model.Role, error) {
	return s.repo.List(ctx)
}

func (s *RoleService) Update(ctx context.Context, id uint64, req *dto.UpdateRoleRequest) error {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Code != "" {
		role.Code = req.Code
	}
	role.Description = req.Description
	if req.Status != nil {
		role.Status = *req.Status
	}
	return s.repo.Update(ctx, role)
}

func (s *RoleService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *RoleService) ListMenus(ctx context.Context) ([]model.Menu, error) {
	return s.repo.ListMenus(ctx)
}

func (s *RoleService) SetRoleMenus(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	return s.repo.SetRoleMenus(ctx, roleID, menuIDs)
}

func (s *RoleService) GetRoleMenus(ctx context.Context, roleID uint64) ([]uint64, error) {
	return s.repo.GetRoleMenus(ctx, roleID)
}

func (s *RoleService) SetUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error {
	return s.repo.SetUserRoles(ctx, userID, roleIDs)
}

func (s *RoleService) GetUserRoles(ctx context.Context, userID uint64) ([]uint64, error) {
	return s.repo.GetUserRoles(ctx, userID)
}
