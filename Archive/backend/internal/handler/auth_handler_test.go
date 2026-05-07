package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cozy-insight-backend/internal/handler"
	"cozy-insight-backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock AuthService
type mockAuthService struct {
	users map[string]*model.User
}

func newMockAuthService() *mockAuthService {
	return &mockAuthService{
		users: make(map[string]*model.User),
	}
}

func (m *mockAuthService) Register(ctx context.Context, user *model.User) error {
	user.ID = "test-user-id"
	m.users[user.Username] = user
	return nil
}

func (m *mockAuthService) Login(ctx context.Context, username, password string) (string, error) {
	return "test-token-12345", nil
}

func (m *mockAuthService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	return &model.User{ID: id, Username: "testuser"}, nil
}

func (m *mockAuthService) UpdatePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	return nil
}

// Tests
func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockAuthService()
	h := handler.NewAuthHandler(mockSvc)

	router := gin.New()
	router.POST("/auth/register", h.Register)

	reqBody := map[string]interface{}{
		"username": "newuser",
		"password": "password123",
		"email":    "user@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-user-id", response.ID)
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockAuthService()
	h := handler.NewAuthHandler(mockSvc)

	router := gin.New()
	router.POST("/auth/login", h.Login)

	reqBody := map[string]interface{}{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
}

func TestAuthHandler_GetCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockAuthService()
	h := handler.NewAuthHandler(mockSvc)

	router := gin.New()
	router.GET("/auth/current", func(c *gin.Context) {
		c.Set("userID", "user-123")
		h.GetCurrentUser(c)
	})

	req, _ := http.NewRequest("GET", "/auth/current", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "user-123", response.ID)
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockAuthService()
	h := handler.NewAuthHandler(mockSvc)

	router := gin.New()
	router.POST("/auth/logout", h.Logout)

	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
