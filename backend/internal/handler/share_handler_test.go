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

// Mock ShareService  
type mockShareService struct {
	shares map[string]*model.Share
}

func newMockShareService() *mockShareService {
	return &mockShareService{
		shares: make(map[string]*model.Share),
	}
}

func (m *mockShareService) Create(ctx context.Context, share *model.Share) error {
	share.ID = "test-share-id"
	share.Token = "test-token-123"
	m.shares[share.ID] = share
	return nil
}

func (m *mockShareService) Delete(ctx context.Context, id string) error {
	delete(m.shares, id)
	return nil
}

func (m *mockShareService) GetByID(ctx context.Context, id string) (*model.Share, error) {
	if share, ok := m.shares[id]; ok {
		return share, nil
	}
	return &model.Share{ID: id, ResourceType: "dashboard"}, nil
}

func (m *mockShareService) List(ctx context.Context, userID string) ([]*model.Share, error) {
	return []*model.Share{
		{ID: "1", ResourceType: "dashboard", Token: "token-1"},
		{ID: "2", ResourceType: "chart", Token: "token-2"},
	}, nil
}

func (m *mockShareService) Validate(ctx context.Context, token, password string) (interface{}, error) {
	return map[string]interface{}{
		"resourceType": "dashboard",
		"resourceId":   "dash-123",
	}, nil
}

// Tests
func TestShareHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockShareService()
	h := handler.NewShareHandler(mockSvc)

	router := gin.New()
	router.POST("/share", h.Create)

	reqBody := map[string]interface{}{
		"resourceType": "dashboard",
		"resourceId":   "dash-123",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/share", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Share
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-share-id", response.ID)
	assert.NotEmpty(t, response.Token)
}

func TestShareHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockShareService()
	h := handler.NewShareHandler(mockSvc)

	router := gin.New()
	router.GET("/share", h.List)

	req, _ := http.NewRequest("GET", "/share", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*model.Share
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
}

func TestShareHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockShareService()
	h := handler.NewShareHandler(mockSvc)

	mockSvc.shares["test-123"] = &model.Share{ID: "test-123"}

	router := gin.New()
	router.DELETE("/share/:id", h.Delete)

	req, _ := http.NewRequest("DELETE", "/share/test-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	_, exists := mockSvc.shares["test-123"]
	assert.False(t, exists)
}

func TestShareHandler_Validate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockShareService()
	h := handler.NewShareHandler(mockSvc)

	router := gin.New()
	router.GET("/share/validate/:token", h.Validate)

	req, _ := http.NewRequest("GET", "/share/validate/test-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
