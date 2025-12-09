package service_test

import (
	"context"
	"errors"
	"testing"

	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthUserRepository是Auth service测试专用mock
type MockAuthUserRepository struct {
	mock.Mock
}

func (m *MockAuthUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		password  string
		email     string
		setupMock func(*MockAuthUserRepository)
		wantErr   bool
	}{
		{
			name:     "Successful registration",
			username: "newuser",
			password: "password123",
			email:    "new@example.com",
			setupMock: func(m *MockAuthUserRepository) {
				m.On("GetByUsername", mock.Anything, "newuser").Return(nil, errors.New("not found"))
				m.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
				m.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "Username already exists",
			username: "existing",
			password: "password123",
			email:    "test@example.com",
			setupMock: func(m *MockAuthUserRepository) {
				m.On("GetByUsername", mock.Anything, "existing").Return(&model.User{ID: "1"}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthUserRepository)
			tt.setupMock(mockRepo)

			authSvc := service.NewAuthService(mockRepo, "test-secret")

			err := authSvc.Register(context.Background(), tt.username, tt.password, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	hashedPassword := "$2a$10$YourHashedPasswordHere" // Placeholder hash

	tests := []struct {
		name      string
		username  string
		password  string
		setupMock func(*MockAuthUserRepository)
		wantErr   bool
	}{
		{
			name:     "User not found",
			username: "nonexistent",
			password: "password",
			setupMock: func(m *MockAuthUserRepository) {
				m.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name:     "Invalid password",
			username: "testuser",
			password: "wrongpassword",
			setupMock: func(m *MockAuthUserRepository) {
				m.On("GetByUsername", mock.Anything, "testuser").Return(&model.User{
					ID:       "1",
					Username: "testuser",
					Password: hashedPassword,
				}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthUserRepository)
			tt.setupMock(mockRepo)

			authSvc := service.NewAuthService(mockRepo, "test-secret")

			token, user, err := authSvc.Login(context.Background(), tt.username, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotNil(t, user)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := new(MockAuthUserRepository)
	secret := "test-secret-key"

	authSvc := service.NewAuthService(mockRepo, secret)

	t.Run("Empty token", func(t *testing.T) {
		claims, err := authSvc.ValidateToken(context.Background(), "")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Invalid token", func(t *testing.T) {
		claims, err := authSvc.ValidateToken(context.Background(), "invalid.token.here")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
