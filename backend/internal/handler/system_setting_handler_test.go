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

// Mock SystemSettingService
type mockSystemSettingService struct {
	settings map[string]*model.SysSetting
}

func newMockSystemSettingService() *mockSystemSettingService {
	return &mockSystemSettingService{
		settings: map[string]*model.SysSetting{
			"system.name": {ID: "1", Type: "system", Key: "system.name", Value: "CozyInsight"},
			"smtp.host":   {ID: "2", Type: "email", Key: "smtp.host", Value: "smtp.example.com"},
		},
	}
}

func (m *mockSystemSettingService) Get(ctx context.Context, key string) (*model.SysSetting, error) {
	if setting, ok := m.settings[key]; ok {
		return setting, nil
	}
	return nil, nil
}

func (m *mockSystemSettingService) Set(ctx context.Context, setting *model.SysSetting) error {
	m.settings[setting.Key] = setting
	return nil
}

func (m *mockSystemSettingService) ListByType(ctx context.Context, settingType string) ([]*model.SysSetting, error) {
	var result []*model.SysSetting
	for _, s := range m.settings {
		if s.Type == settingType {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockSystemSettingService) Delete(ctx context.Context, key string) error {
	delete(m.settings, key)
	return nil
}

// Tests
func TestSystemSettingHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockSystemSettingService()
	h := handler.NewSystemSettingHandler(mockSvc)

	router := gin.New()
	router.GET("/setting/:key", h.Get)

	req, _ := http.NewRequest("GET", "/setting/system.name", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.SysSetting
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "CozyInsight", response.Value)
}

func TestSystemSettingHandler_Set(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockSystemSettingService()
	h := handler.NewSystemSettingHandler(mockSvc)

	router := gin.New()
	router.POST("/setting", h.Set)

	reqBody := map[string]interface{}{
		"type":  "system",
		"key":   "new.setting",
		"value": "test value",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/setting", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, mockSvc.settings["new.setting"])
}

func TestSystemSettingHandler_ListByType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockSystemSettingService()
	h := handler.NewSystemSettingHandler(mockSvc)

	router := gin.New()
	router.GET("/setting/type/:type", h.ListByType)

	req, _ := http.NewRequest("GET", "/setting/type/system", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*model.SysSetting
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.GreaterOrEqual(t, len(response), 1)
}

func TestSystemSettingHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockSystemSettingService()
	h := handler.NewSystemSettingHandler(mockSvc)

	router := gin.New()
	router.DELETE("/setting/:key", h.Delete)

	req, _ := http.NewRequest("DELETE", "/setting/smtp.host", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	_, exists := mockSvc.settings["smtp.host"]
	assert.False(t, exists)
}
