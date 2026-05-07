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

// Mock DashboardService
type mockDashboardService struct {
	dashboards map[string]*model.Dashboard
}

func newMockDashboardService() *mockDashboardService {
	return &mockDashboardService{
		dashboards: make(map[string]*model.Dashboard),
	}
}

func (m *mockDashboardService) Create(ctx context.Context, dashboard *model.Dashboard) error {
	dashboard.ID = "test-dashboard-id"
	m.dashboards[dashboard.ID] = dashboard
	return nil
}

func (m *mockDashboardService) Update(ctx context.Context, dashboard *model.Dashboard) error {
	m.dashboards[dashboard.ID] = dashboard
	return nil
}

func (m *mockDashboardService) Delete(ctx context.Context, id string) error {
	delete(m.dashboards, id)
	return nil
}

func (m *mockDashboardService) GetByID(ctx context.Context, id string) (*model.Dashboard, error) {
	if dashboard, ok := m.dashboards[id]; ok {
		return dashboard, nil
	}
	return &model.Dashboard{ID: id, Name: "Test Dashboard"}, nil
}

func (m *mockDashboardService) List(ctx context.Context) ([]*model.Dashboard, error) {
	return []*model.Dashboard{
		{ID: "1", Name: "Dashboard 1"},
		{ID: "2", Name: "Dashboard 2"},
	}, nil
}

func (m *mockDashboardService) Publish(ctx context.Context, id string) error {
	if dashboard, ok := m.dashboards[id]; ok {
		dashboard.Status = "published"
	}
	return nil
}

func (m *mockDashboardService) Unpublish(ctx context.Context, id string) error {
	if dashboard, ok := m.dashboards[id]; ok {
		dashboard.Status = "draft"
	}
	return nil
}

func (m *mockDashboardService) SaveComponents(ctx context.Context, dashboardID string, components interface{}) error {
	return nil
}

func (m *mockDashboardService) GetComponents(ctx context.Context, dashboardID string) (interface{}, error) {
	return []map[string]interface{}{
		{"id": "1", "type": "chart"},
	}, nil
}

func (m *mockDashboardService) GetPublished(ctx context.Context, id string) (*model.Dashboard, error) {
	return &model.Dashboard{ID: id, Name: "Published Dashboard", Status: "published"}, nil
}

// Tests
func TestDashboardHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockDashboardService()
	h := handler.NewDashboardHandler(mockSvc)

	router := gin.New()
	router.POST("/dashboard", h.Create)

	reqBody := map[string]interface{}{
		"name":     "Test Dashboard",
		"nodeType": "dashboard",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/dashboard", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Dashboard
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-dashboard-id", response.ID)
}

func TestDashboardHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockDashboardService()
	h := handler.NewDashboardHandler(mockSvc)

	router := gin.New()
	router.GET("/dashboard", h.List)

	req, _ := http.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*model.Dashboard
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
}

func TestDashboardHandler_Publish(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockDashboardService()
	h := handler.NewDashboardHandler(mockSvc)

	mockSvc.dashboards["test-123"] = &model.Dashboard{
		ID:     "test-123",
		Name:   "Test",
		Status: "draft",
	}

	router := gin.New()
	router.POST("/dashboard/:id/publish", h.Publish)

	req, _ := http.NewRequest("POST", "/dashboard/test-123/publish", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "published", mockSvc.dashboards["test-123"].Status)
}

func TestDashboardHandler_Unpublish(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockDashboardService()
	h := handler.NewDashboardHandler(mockSvc)

	mockSvc.dashboards["test-123"] = &model.Dashboard{
		ID:     "test-123",
		Name:   "Test",
		Status: "published",
	}

	router := gin.New()
	router.POST("/dashboard/:id/unpublish", h.Unpublish)

	req, _ := http.NewRequest("POST", "/dashboard/test-123/unpublish", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "draft", mockSvc.dashboards["test-123"].Status)
}
