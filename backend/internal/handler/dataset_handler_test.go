package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/engine"
	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
)

type testConnector struct {
	columns []engine.ColumnInfo
	data    []map[string]interface{}
}

func (m *testConnector) Connect(configJSON string) error { return nil }
func (m *testConnector) Close() error                      { return nil }
func (m *testConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	return m.data, nil
}
func (m *testConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]engine.ColumnInfo, error) {
	return m.columns, nil
}

func setupDatasetHandler(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	dsRepo := repository.NewDatasourceRepository(sqlxDB)
	rowPermRepo := repository.NewRowPermissionRepository(sqlxDB)
	repo := repository.NewDatasetRepository(sqlxDB)
	svc := service.NewDatasetService(repo, dsRepo, rowPermRepo)
	svc.SetConnectorFactory(func(string) (engine.DatasourceConnector, error) {
		return &testConnector{
			columns: []engine.ColumnInfo{{Name: "id", Type: "INT", Length: 11}},
			data:    []map[string]interface{}{{ "id": 1 }},
		}, nil
	})
	h := NewDatasetHandler(svc)

	r := gin.New()
	r.GET("/dataset", h.List)
	r.POST("/dataset", h.Create)
	r.GET("/dataset/:id", h.Get)
	r.PUT("/dataset/:id", h.Update)
	r.DELETE("/dataset/:id", h.Delete)
	r.POST("/dataset/:id/sync-fields", h.SyncFields)
	r.GET("/dataset/:id/preview", h.Preview)
	return r, mock
}

func TestDatasetHandler_Create(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	mock.ExpectExec("INSERT INTO datasets").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(dto.CreateDatasetRequest{
		Name:         "Test Dataset",
		DatasourceID: 1,
		DatabaseName: "test_db",
		TableName:    "test_table",
		Type:         "table",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dataset", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasetHandler_Get(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Test Dataset", 1, "test_db", "test_table", "table", 0, 1, 1, now, now, nil,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasetHandler_Get_InvalidID(t *testing.T) {
	r, _ := setupDatasetHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasetHandler_List(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE deleted_at IS NULL").
		WillReturnRows(sqlmock.NewRows(columns))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasetHandler_Update(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	columns := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			1, "Old Name", 1, "test_db", "test_table", "table", 0, 1, 1, now, now, nil,
		))

	mock.ExpectExec("UPDATE datasets SET").
		WillReturnResult(sqlmock.NewResult(0, 1))

	body, _ := json.Marshal(dto.UpdateDatasetRequest{
		Name: "Updated Dataset",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/dataset/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasetHandler_Delete(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	mock.ExpectExec("UPDATE datasets SET deleted_at = NOW").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/dataset/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasetHandler_SyncFields(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	dsColumns := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	dsNow := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsColumns).AddRow(
			1, "Test Dataset", 1, "test_db", "test_table", "table", 0, 1, 1, dsNow, dsNow, nil,
		))

	datasourceColumns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceColumns).AddRow(
			1, "MySQL", "mysql", "{}", 1, 1, dsNow, dsNow, nil,
		))

	mock.ExpectExec("DELETE FROM dataset_fields WHERE dataset_id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("INSERT INTO dataset_fields").
		WillReturnResult(sqlmock.NewResult(1, 3))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dataset/1/sync-fields", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasetHandler_Preview(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	dsColumns := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	dsNow := time.Now()
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsColumns).AddRow(
			1, "Test Dataset", 1, "test_db", "test_table", "table", 0, 1, 1, dsNow, dsNow, nil,
		))

	datasourceColumns := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceColumns).AddRow(
			1, "MySQL", "mysql", "{}", 1, 1, dsNow, dsNow, nil,
		))

	fieldColumns := []string{"id", "dataset_id", "name", "type", "de_type", "length", "precision", "scale", "origin_name", "created_at", "updated_at"}
	mock.ExpectQuery("SELECT \\* FROM dataset_fields WHERE dataset_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(fieldColumns))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset/1/preview?limit=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatasetHandler_Create_InvalidBody(t *testing.T) {
	r, _ := setupDatasetHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dataset", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasetHandler_Get_NotFound(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDatasetHandler_List_DBError(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	mock.ExpectQuery("SELECT \\* FROM datasets WHERE deleted_at IS NULL").
		WillReturnError(sql.ErrConnDone)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDatasetHandler_Update_NotFound(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	body, _ := json.Marshal(dto.UpdateDatasetRequest{Name: "Updated"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/dataset/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasetHandler_Update_InvalidBody(t *testing.T) {
	r, _ := setupDatasetHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/dataset/1", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasetHandler_Delete_NotFound(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	mock.ExpectExec("UPDATE datasets SET deleted_at = NOW").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/dataset/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasetHandler_Preview_NotFound(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dataset/1/preview?limit=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDatasetHandler_SyncFields_NotFound(t *testing.T) {
	r, mock := setupDatasetHandler(t)

	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dataset/1/sync-fields", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
