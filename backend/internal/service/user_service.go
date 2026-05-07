package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
	_, err := s.repo.FindByUsername(ctx, req.Username)
	if err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	isAdmin := int8(0)
	if req.IsAdmin {
		isAdmin = 1
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		NickName:     req.NickName,
		Phone:        req.Phone,
		Status:       req.Status,
		IsAdmin:      isAdmin,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) Update(ctx context.Context, id uint64, req *dto.UpdateUserRequest) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.NickName != "" {
		user.NickName = req.NickName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	user.Status = req.Status
	if req.IsAdmin {
		user.IsAdmin = 1
	} else {
		user.IsAdmin = 0
	}

	return s.repo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) ChangePassword(ctx context.Context, id uint64, req *dto.ChangePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("old password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password failed: %w", err)
	}

	return s.repo.UpdatePassword(ctx, id, string(hash))
}
