package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cozy-insight-backend/internal/handler"
	"cozy-insight-backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock OperLogService
type mockOperLogService struct {
	logs []*model.SysOperLog
}

func newMockOperLogService() *mockOperLogService {
	return &mockOperLogService{
		logs: []*model.SysOperLog{
			{ID: "1", Module: "用户", Action: "登录", IP: "127.0.0.1", CreatedAt: time.Now()},
			{ID: "2", Module: "数据集", Action: "创建", IP: "192.168.1.1", CreatedAt: time.Now()},
		},
	}
}

func (m *mockOperLogService) List(ctx context.Context, limit int) ([]*model.SysOperLog, error) {
	if limit > 0 && limit < len(m.logs) {
		return m.logs[:limit], nil
	}
	return m.logs, nil
}

func (m *mockOperLogService) CleanOldLogs(ctx context.Context, days int) error {
	return nil
}

// Tests
func TestOperLogHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockOperLogService()
	h := handler.NewOperLogHandler(mockSvc)

	router := gin.New()
	router.GET("/log", h.List)

	req, _ := http.NewRequest("GET", "/log?limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*model.SysOperLog
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
}

func TestOperLogHandler_Clean(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockOperLogService()
	h := handler.NewOperLogHandler(mockSvc)

	router := gin.New()
	router.POST("/log/clean", h.Clean)

	reqBody := map[string]interface{}{
		"days": 30,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/log/clean", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
