package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"cozy-insight-backend/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock ExportService
type mockExportService struct{}

func (m *mockExportService) ExportToExcel(ctx context.Context, data interface{}) ([]byte, error) {
	return []byte("mock-excel-data"), nil
}

func (m *mockExportService) ExportToCSV(ctx context.Context, data interface{}) ([]byte, error) {
	return []byte("mock-csv-data"), nil
}

func (m *mockExportService) ExportChartImage(ctx context.Context, chartID string) ([]byte, error) {
	return []byte("mock-image-data"), nil
}

// Tests
func TestExportHandler_ExportDatasetToExcel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockExportSvc := &mockExportService{}
	mockDatasetSvc := newMockDatasetService()
	h := handler.NewExportHandler(mockExportSvc, mockDatasetSvc, nil)

	router := gin.New()
	router.GET("/export/dataset/:id/excel", h.ExportDatasetToExcel)

	req, _ := http.NewRequest("GET", "/export/dataset/test-123/excel", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application")
}

func TestExportHandler_ExportDatasetToCSV(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockExportSvc := &mockExportService{}
	mockDatasetSvc := newMockDatasetService()
	h := handler.NewExportHandler(mockExportSvc, mockDatasetSvc, nil)

	router := gin.New()
	router.GET("/export/dataset/:id/csv", h.ExportDatasetToCSV)

	req, _ := http.NewRequest("GET", "/export/dataset/test-123/csv", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExportHandler_ExportChartData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockExportSvc := &mockExportService{}
	h := handler.NewExportHandler(mockExportSvc, nil, nil)

	router := gin.New()
	router.GET("/export/chart/:id", h.ExportChartData)

	req, _ := http.NewRequest("GET", "/export/chart/test-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExportHandler_ExportChartImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockExportSvc := &mockExportService{}
	h := handler.NewExportHandler(mockExportSvc, nil, nil)

	router := gin.New()
	router.GET("/export/chart/:id/image", h.ExportChartImage)

	req, _ := http.NewRequest("GET", "/export/chart/test-123/image", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
