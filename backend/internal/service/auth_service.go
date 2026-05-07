package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
	"cozy-insight/pkg/jwt"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	jwtManager  *jwt.Manager
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *jwt.Manager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	_, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil {
		return fmt.Errorf("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password failed: %w", err)
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		NickName:     req.NickName,
		Status:       1,
		IsAdmin:      0,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if user.Status != 1 {
		return nil, fmt.Errorf("user is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	token, err := s.jwtManager.Generate(user.ID, user.Username, user.IsAdmin == 1)
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}

	return &dto.LoginResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		NickName: user.NickName,
		IsAdmin:  user.IsAdmin == 1,
	}, nil
}
