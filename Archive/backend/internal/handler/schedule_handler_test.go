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

// Mock ScheduleService
type mockScheduleService struct {
	tasks map[string]*model.ScheduleTask
}

func newMockScheduleService() *mockScheduleService {
	return &mockScheduleService{
		tasks: make(map[string]*model.ScheduleTask),
	}
}

func (m *mockScheduleService) Create(ctx context.Context, task *model.ScheduleTask) error {
	task.ID = "test-task-id"
	m.tasks[task.ID] = task
	return nil
}

func (m *mockScheduleService) Update(ctx context.Context, task *model.ScheduleTask) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockScheduleService) Delete(ctx context.Context, id string) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockScheduleService) GetByID(ctx context.Context, id string) (*model.ScheduleTask, error) {
	if task, ok := m.tasks[id]; ok {
		return task, nil
	}
	return &model.ScheduleTask{ID: id, Name: "Test Task"}, nil
}

func (m *mockScheduleService) List(ctx context.Context) ([]*model.ScheduleTask, error) {
	return []*model.ScheduleTask{
		{ID: "1", Name: "Task 1", Enabled: true},
		{ID: "2", Name: "Task 2", Enabled: false},
	}, nil
}

func (m *mockScheduleService) Enable(ctx context.Context, id string) error {
	if task, ok := m.tasks[id]; ok {
		task.Enabled = true
	}
	return nil
}

func (m *mockScheduleService) Disable(ctx context.Context, id string) error {
	if task, ok := m.tasks[id]; ok {
		task.Enabled = false
	}
	return nil
}

func (m *mockScheduleService) Execute(ctx context.Context, id string) error {
	return nil
}

func (m *mockScheduleService) Start() error { return nil }
func (m *mockScheduleService) Stop()        {}

// Tests
func TestScheduleHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockScheduleService()
	h := handler.NewScheduleHandler(mockSvc)

	router := gin.New()
	router.POST("/schedule", h.Create)

	reqBody := map[string]interface{}{
		"name":     "Test Task",
		"type":     "email_report",
		"cronExpr": "0 0 * * *",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/schedule", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.ScheduleTask
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-task-id", response.ID)
}

func TestScheduleHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockScheduleService()
	h := handler.NewScheduleHandler(mockSvc)

	router := gin.New()
	router.GET("/schedule", h.List)

	req, _ := http.NewRequest("GET", "/schedule", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*model.ScheduleTask
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
}

func TestScheduleHandler_Enable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockScheduleService()
	h := handler.NewScheduleHandler(mockSvc)

	mockSvc.tasks["test-123"] = &model.ScheduleTask{ID: "test-123", Enabled: false}

	router := gin.New()
	router.POST("/schedule/:id/enable", h.Enable)

	req, _ := http.NewRequest("POST", "/schedule/test-123/enable", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, mockSvc.tasks["test-123"].Enabled)
}

func TestScheduleHandler_Disable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockScheduleService()
	h := handler.NewScheduleHandler(mockSvc)

	mockSvc.tasks["test-123"] = &model.ScheduleTask{ID: "test-123", Enabled: true}

	router := gin.New()
	router.POST("/schedule/:id/disable", h.Disable)

	req, _ := http.NewRequest("POST", "/schedule/test-123/disable", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, mockSvc.tasks["test-123"].Enabled)
}

func TestScheduleHandler_Execute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockScheduleService()
	h := handler.NewScheduleHandler(mockSvc)

	router := gin.New()
	router.POST("/schedule/:id/execute", h.Execute)

	req, _ := http.NewRequest("POST", "/schedule/test-123/execute", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
