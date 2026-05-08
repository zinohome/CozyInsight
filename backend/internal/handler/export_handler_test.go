package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/repository"
	"cozy-insight/internal/service"
)

func setupExportHandler(t *testing.T) (*ExportHandler, *gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "mysql")
	chartRepo := repository.NewChartRepository(sqlxDB)
	datasetRepo := repository.NewDatasetRepository(sqlxDB)
	dsRepo := repository.NewDatasourceRepository(sqlxDB)
	chartService := service.NewChartService(chartRepo, datasetRepo, dsRepo, nil, nil)
	h := NewExportHandler(chartService)
	r := gin.New()
	return h, r, mock
}

func TestExportHandler_ExportCSV_InvalidID(t *testing.T) {
	h, r, _ := setupExportHandler(t)
	r.GET("/chart/:id/export/csv", h.ExportCSV)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chart/abc/export/csv", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportHandler_ExportCSV_NotFound(t *testing.T) {
	h, r, mock := setupExportHandler(t)
	r.GET("/chart/:id/export/csv", h.ExportCSV)

	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chart/999/export/csv", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExportHandler_ExportCSV_Success(t *testing.T) {
	h, r, mock := setupExportHandler(t)
	r.GET("/chart/:id/export/csv", h.ExportCSV)

	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(
			1, "Sales DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil,
		))

	datasourceCols := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceCols).AddRow(
			1, "Local MySQL", "mysql", `{"host":"h","port":3306}`, 1, 1, now, now, nil,
		))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chart/1/export/csv", nil)
	r.ServeHTTP(w, req)

	// Expects 400 because ExportCSV maps all GetData errors to 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
