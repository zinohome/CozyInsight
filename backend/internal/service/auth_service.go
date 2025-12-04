package service

import (
	"context"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	jwtutil "cozy-insight-backend/pkg/jwt"
	"cozy-insight-backend/pkg/logger"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*model.User, error)
	Login(ctx context.Context, username, password string) (string, *model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
}

type authService struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewAuthService(repo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	// 验证输入
	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("username, email and password are required")
	}

	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	// 检查用户名是否已存在
	existingUser, _ := s.repo.GetByUsername(ctx, username)
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// 检查邮箱是否已存在
	existingEmail, _ := s.repo.GetByEmail(ctx, email)
	if existingEmail != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password")
	}

	// 创建用户
	user := &model.User{
		ID:         uuid.New().String(),
		Username:   username,
		Email:      email,
		Password:   string(hashedPassword),
		Role:       "user",
		CreateTime: time.Now().UnixMilli(),
		UpdateTime: time.Now().UnixMilli(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Log.Error("failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Log.Info("user registered successfully", zap.String("username", username))
	return user, nil
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, username, password string) (string, *model.User, error) {
	// 验证输入
	if username == "" || password == "" {
		return "", nil, fmt.Errorf("username and password are required")
	}

	// 查找用户
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		logger.Log.Warn("user not found", zap.String("username", username))
		return "", nil, fmt.Errorf("invalid username or password")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.Log.Warn("invalid password", zap.String("username", username))
		return "", nil, fmt.Errorf("invalid username or password")
	}

	// 生成 Token
	token, err := jwtutil.GenerateToken(user.ID, user.Username, user.Role, s.jwtSecret)
	if err != nil {
		logger.Log.Error("failed to generate token", zap.Error(err))
		return "", nil, fmt.Errorf("failed to generate token")
	}

	logger.Log.Info("user logged in successfully", zap.String("username", username))
	return token, user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *authService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}
