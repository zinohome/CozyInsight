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

// Mock PermissionService
type mockPermissionService struct{}

func (m *mockPermissionService) List(ctx context.Context) ([]*model.Permission, error) {
	return []*model.Permission{
		{ID: "1", Name: "read", Resource: "dashboard"},
		{ID: "2", Name: "write", Resource: "dataset"},
	}, nil
}

func (m *mockPermissionService) GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error) {
	return []*model.Permission{
		{ID: "1", Name: "read", Resource: "dashboard"},
	}, nil
}

func (m *mockPermissionService) GrantToRole(ctx context.Context, roleID, permissionID string) error {
	return nil
}

func (m *mockPermissionService) RevokeFromRole(ctx context.Context, roleID, permissionID string) error {
	return nil
}

func (m *mockPermissionService) GrantResourcePermission(ctx context.Context, perm *model.ResourcePermission) error {
	return nil
}

func (m *mockPermissionService) GetResourcePermissions(ctx context.Context, resourceType, resourceID string) ([]*model.ResourcePermission, error) {
	return []*model.ResourcePermission{
		{ID: "1", ResourceType: "dashboard", ResourceID: "dash-1"},
	}, nil
}

func (m *mockPermissionService) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	return true, nil
}

// Tests
func TestPermissionHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockPermissionService{}
	h := handler.NewPermissionHandler(mockSvc)

	router := gin.New()
	router.GET("/permission", h.List)

	req, _ := http.NewRequest("GET", "/permission", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*model.Permission
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
}

func TestPermissionHandler_GetRolePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockPermissionService{}
	h := handler.NewPermissionHandler(mockSvc)

	router := gin.New()
	router.GET("/permission/role/:roleId", h.GetRolePermissions)

	req, _ := http.NewRequest("GET", "/permission/role/role-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionHandler_GrantToRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockPermissionService{}
	h := handler.NewPermissionHandler(mockSvc)

	router := gin.New()
	router.POST("/permission/role/:roleId/grant", h.GrantToRole)

	reqBody := map[string]interface{}{
		"permissionId": "perm-123",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/permission/role/role-123/grant", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionHandler_RevokeFromRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockPermissionService{}
	h := handler.NewPermissionHandler(mockSvc)

	router := gin.New()
	router.POST("/permission/role/:roleId/revoke", h.RevokeFromRole)

	reqBody := map[string]interface{}{
		"permissionId": "perm-123",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/permission/role/role-123/revoke", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionHandler_CheckPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockPermissionService{}
	h := handler.NewPermissionHandler(mockSvc)

	router := gin.New()
	router.GET("/permission/check", h.CheckPermission)

	req, _ := http.NewRequest("GET", "/permission/check?resource=dashboard&action=read", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
