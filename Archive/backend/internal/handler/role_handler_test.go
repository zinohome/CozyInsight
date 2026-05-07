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

// Mock RoleService
type mockRoleService struct {
	roles map[string]*model.Role
}

func newMockRoleService() *mockRoleService {
	return &mockRoleService{
		roles: make(map[string]*model.Role),
	}
}

func (m *mockRoleService) Create(ctx context.Context, role *model.Role) error {
	role.ID = "test-role-id"
	m.roles[role.ID] = role
	return nil
}

func (m *mockRoleService) Update(ctx context.Context, role *model.Role) error {
	m.roles[role.ID] = role
	return nil
}

func (m *mockRoleService) Delete(ctx context.Context, id string) error {
	delete(m.roles, id)
	return nil
}

func (m *mockRoleService) GetByID(ctx context.Context, id string) (*model.Role, error) {
	if role, ok := m.roles[id]; ok {
		return role, nil
	}
	return &model.Role{ID: id, Name: "Test Role"}, nil
}

func (m *mockRoleService) List(ctx context.Context) ([]*model.Role, error) {
	var result []*model.Role
	for _, role := range m.roles {
		result = append(result, role)
	}
	if len(result) == 0 {
		return []*model.Role{
			{ID: "1", Name: "Admin", Type: "system"},
			{ID: "2", Name: "User", Type: "custom"},
		}, nil
	}
	return result, nil
}

func (m *mockRoleService) AssignToUser(ctx context.Context, roleID, userID string) error {
	return nil
}

func (m *mockRoleService) RemoveFromUser(ctx context.Context, roleID, userID string) error {
	return nil
}

func (m *mockRoleService) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	return []*model.Role{
		{ID: "1", Name: "Admin"},
	}, nil
}

// Tests
func TestRoleHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockRoleService()
	h := handler.NewRoleHandler(mockSvc)

	router := gin.New()
	router.POST("/role", h.Create)

	reqBody := map[string]interface{}{
		"name":        "Test Role",
		"description": "Test Description",
		"type":        "custom",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Role
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-role-id", response.ID)
}

func TestRoleHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockRoleService()
	h := handler.NewRoleHandler(mockSvc)

	router := gin.New()
	router.GET("/role", h.List)

	req, _ := http.NewRequest("GET", "/role", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*model.Role
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
}

func TestRoleHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockRoleService()
	h := handler.NewRoleHandler(mockSvc)

	router := gin.New()
	router.GET("/role/:id", h.Get)

	req, _ := http.NewRequest("GET", "/role/test-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockRoleService()
	h := handler.NewRoleHandler(mockSvc)

	mockSvc.roles["test-123"] = &model.Role{
		ID:   "test-123",
		Name: "Old Role",
	}

	router := gin.New()
	router.PUT("/role/:id", h.Update)

	reqBody := map[string]interface{}{
		"name": "Updated Role",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/role/test-123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockRoleService()
	h := handler.NewRoleHandler(mockSvc)

	mockSvc.roles["test-123"] = &model.Role{ID: "test-123"}

	router := gin.New()
	router.DELETE("/role/:id", h.Delete)

	req, _ := http.NewRequest("DELETE", "/role/test-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	_, exists := mockSvc.roles["test-123"]
	assert.False(t, exists)
}

func TestRoleHandler_AssignToUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockRoleService()
	h := handler.NewRoleHandler(mockSvc)

	router := gin.New()
	router.POST("/role/:id/assign", h.AssignToUser)

	reqBody := map[string]interface{}{
		"userId": "user-123",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/role/role-123/assign", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
