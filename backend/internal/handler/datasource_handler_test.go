package handler_test

import (
	"bytes"
	"cozy-insight-backend/internal/handler"
	"cozy-insight-backend/internal/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock Service
type mockDatasourceService struct{}

func (m *mockDatasourceService) Create(ctx context.Context, ds *model.Datasource) error {
	ds.ID = "test-id"
	return nil
}

func (m *mockDatasourceService) List(ctx context.Context) ([]*model.Datasource, error) {
	return []*model.Datasource{
		{ID: "1", Name: "MySQL", Type: "mysql"},
		{ID: "2", Name: "PostgreSQL", Type: "postgresql"},
	}, nil
}

func (m *mockDatasourceService) GetByID(ctx context.Context, id string) (*model.Datasource, error) {
	return &model.Datasource{ID: id, Name: "Test", Type: "mysql"}, nil
}

func (m *mockDatasourceService) Update(ctx context.Context, ds *model.Datasource) error {
	return nil
}

func (m *mockDatasourceService) Delete(ctx context.Context, id string) error {
	return nil
}

// Tests
func TestDatasourceHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockService := &mockDatasourceService{}
	handler := handler.NewDatasourceHandler(mockService)
	
	router := gin.New()
	router.POST("/datasource", handler.Create)
	
	reqBody := map[string]interface{}{
		"name":   "Test MySQL",
		"type":   "mysql",
		"config": `{"host":"localhost"}`,
	}
	body, _ := json.Marshal(reqBody)
	
	req, _ := http.NewRequest("POST", "/datasource", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response model.Datasource
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-id", response.ID)
}

func TestDatasourceHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockService := &mockDatasourceService{}
	handler := handler.NewDatasourceHandler(mockService)
	
	router := gin.New()
	router.GET("/datasource", handler.List)
	
	req, _ := http.NewRequest("GET", "/datasource", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response []*model.Datasource
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
	assert.Equal(t, "MySQL", response[0].Name)
}

func TestDatasourceHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockService := &mockDatasourceService{}
	handler := handler.NewDatasourceHandler(mockService)
	
	router := gin.New()
	router.GET("/datasource/:id", handler.Get)
	
	req, _ := http.NewRequest("GET", "/datasource/test-123", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response model.Datasource
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-123", response.ID)
}
