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

// Mock DatasetService
type mockDatasetService struct {
	datasets map[string]*model.DatasetTable
}

func newMockDatasetService() *mockDatasetService {
	return &mockDatasetService{
		datasets: make(map[string]*model.DatasetTable),
	}
}

func (m *mockDatasetService) Create(ctx context.Context, dataset *model.DatasetTable) error {
	dataset.ID = "test-dataset-id"
	m.datasets[dataset.ID] = dataset
	return nil
}

func (m *mockDatasetService) Update(ctx context.Context, dataset *model.DatasetTable) error {
	m.datasets[dataset.ID] = dataset
	return nil
}

func (m *mockDatasetService) Delete(ctx context.Context, id string) error {
	delete(m.datasets, id)
	return nil
}

func (m *mockDatasetService) GetByID(ctx context.Context, id string) (*model.DatasetTable, error) {
	if dataset, ok := m.datasets[id]; ok {
		return dataset, nil
	}
	return &model.DatasetTable{ID: id, Name: "Test Dataset", Type: "db"}, nil
}

func (m *mockDatasetService) List(ctx context.Context) ([]*model.DatasetTable, error) {
	var result []*model.DatasetTable
	for _, dataset := range m.datasets {
		result = append(result, dataset)
	}
	if len(result) == 0 {
		return []*model.DatasetTable{
			{ID: "1", Name: "Dataset 1", Type: "db"},
			{ID: "2", Name: "Dataset 2", Type: "sql"},
		}, nil
	}
	return result, nil
}

func (m *mockDatasetService) GetPreviewData(ctx context.Context, id string, limit int) (interface{}, error) {
	return []map[string]interface{}{
		{"id": 1, "name": "Row 1"},
		{"id": 2, "name": "Row 2"},
	}, nil
}

func (m *mockDatasetService) SyncFields(ctx context.Context, datasetID string) error {
	return nil
}

// Tests
func TestDatasetHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockDatasetService()
	h := handler.NewDatasetHandler(mockSvc)

	router := gin.New()
	router.POST("/dataset", h.Create)

	reqBody := map[string]interface{}{
		"name": "Test Dataset",
		"type": "db",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/dataset", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.DatasetTable
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-dataset-id", response.ID)
}

func TestDatasetHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockDatasetService()
	h := handler.NewDatasetHandler(mockSvc)

	router := gin.New()
	router.GET("/dataset", h.List)

	req, _ := http.NewRequest("GET", "/dataset", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*model.DatasetTable
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
}

func TestDatasetHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockDatasetService()
	h := handler.NewDatasetHandler(mockSvc)

	router := gin.New()
	router.GET("/dataset/:id", h.Get)

	req, _ := http.NewRequest("GET", "/dataset/test-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasetHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockDatasetService()
	h := handler.NewDatasetHandler(mockSvc)

	mockSvc.datasets["test-123"] = &model.DatasetTable{
		ID:   "test-123",
		Name: "Old Name",
		Type: "db",
	}

	router := gin.New()
	router.PUT("/dataset/:id", h.Update)

	reqBody := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/dataset/test-123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasetHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := newMockDatasetService()
	h := handler.NewDatasetHandler(mockSvc)

	mockSvc.datasets["test-123"] = &model.DatasetTable{ID: "test-123"}

	router := gin.New()
	router.DELETE("/dataset/:id", h.Delete)

	req, _ := http.NewRequest("DELETE", "/dataset/test-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	_, exists := mockSvc.datasets["test-123"]
	assert.False(t, exists)
}
