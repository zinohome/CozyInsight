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

// Mock ChartService
type mockChartService struct {
	charts map[string]*model.Chart
}

func newMockChartService() *mockChartService {
	return &mockChartService{
		charts: make(map[string]*model.Chart),
	}
}

func (m *mockChartService) Create(ctx context.Context, chart *model.Chart) error {
	chart.ID = "test-chart-id"
	m.charts[chart.ID] = chart
	return nil
}

func (m *mockChartService) Update(ctx context.Context, chart *model.Chart) error {
	m.charts[chart.ID] = chart
	return nil
}

func (m *mockChartService) Delete(ctx context.Context, id string) error {
	delete(m.charts, id)
	return nil
}

func (m *mockChartService) GetByID(ctx context.Context, id string) (*model.Chart, error) {
	if chart, ok := m.charts[id]; ok {
		return chart, nil
	}
	return &model.Chart{ID: id, Name: "Test Chart", Type: "bar"}, nil
}

func (m *mockChartService) List(ctx context.Context, userID string) ([]*model.Chart, error) {
	var result []*model.Chart
	for _, chart := range m.charts {
		result = append(result, chart)
	}
	if len(result) == 0 {
		// Return default test data
		return []*model.Chart{
			{ID: "1", Name: "Chart 1", Type: "bar"},
			{ID: "2", Name: "Chart 2", Type: "line"},
		}, nil
	}
	return result, nil
}

// Mock ChartDataService
type mockChartDataService struct{}

func (m *mockChartDataService) GetChartData(ctx context.Context, chartID string, params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"data": []map[string]interface{}{
			{"x": "A", "y": 10},
			{"x": "B", "y": 20},
		},
	}, nil
}

// Tests
func TestChartHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChartSvc := newMockChartService()
	mockDataSvc := &mockChartDataService{}
	h := handler.NewChartHandler(mockChartSvc, mockDataSvc)

	router := gin.New()
	router.POST("/chart", h.Create)

	reqBody := map[string]interface{}{
		"name":         "Test Bar Chart",
		"type":         "bar",
		"datasetTable": "test_table",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/chart", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Chart
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-chart-id", response.ID)
	assert.Equal(t, "Test Bar Chart", response.Name)
}

func TestChartHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChartSvc := newMockChartService()
	mockDataSvc := &mockChartDataService{}
	h := handler.NewChartHandler(mockChartSvc, mockDataSvc)

	router := gin.New()
	router.GET("/chart", h.List)

	req, _ := http.NewRequest("GET", "/chart", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*model.Chart
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
	assert.Equal(t, "Chart 1", response[0].Name)
}

func TestChartHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChartSvc := newMockChartService()
	mockDataSvc := &mockChartDataService{}
	h := handler.NewChartHandler(mockChartSvc, mockDataSvc)

	router := gin.New()
	router.GET("/chart/:id", h.Get)

	req, _ := http.NewRequest("GET", "/chart/test-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Chart
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-123", response.ID)
}

func TestChartHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChartSvc := newMockChartService()
	mockDataSvc := &mockChartDataService{}
	h := handler.NewChartHandler(mockChartSvc, mockDataSvc)

	// Pre-create a chart
	mockChartSvc.charts["test-123"] = &model.Chart{
		ID:   "test-123",
		Name: "Old Name",
		Type: "bar",
	}

	router := gin.New()
	router.PUT("/chart/:id", h.Update)

	reqBody := map[string]interface{}{
		"name": "Updated Name",
		"type": "line",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/chart/test-123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	updated := mockChartSvc.charts["test-123"]
	assert.Equal(t, "Updated Name", updated.Name)
}

func TestChartHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChartSvc := newMockChartService()
	mockDataSvc := &mockChartDataService{}
	h := handler.NewChartHandler(mockChartSvc, mockDataSvc)

	// Pre-create a chart
	mockChartSvc.charts["test-123"] = &model.Chart{ID: "test-123"}

	router := gin.New()
	router.DELETE("/chart/:id", h.Delete)

	req, _ := http.NewRequest("DELETE", "/chart/test-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	_, exists := mockChartSvc.charts["test-123"]
	assert.False(t, exists)
}

func TestChartHandler_GetData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockChartSvc := newMockChartService()
	mockDataSvc := &mockChartDataService{}
	h := handler.NewChartHandler(mockChartSvc, mockDataSvc)

	router := gin.New()
	router.GET("/chart/:id/data", h.GetData)

	req, _ := http.NewRequest("GET", "/chart/test-123/data", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])
}
