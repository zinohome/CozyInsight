package handler

import (
	"cozy-insight-backend/internal/service"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	exportService  service.ExportService
	datasetService service.DatasetService
	chartService   service.ChartDataService
}

func NewExportHandler(
	exportService service.ExportService,
	datasetService service.DatasetService,
	chartService service.ChartDataService,
) *ExportHandler {
	return &ExportHandler{
		exportService:  exportService,
		datasetService: datasetService,
		chartService:   chartService,
	}
}

// ExportDataset 导出数据集
func (h *ExportHandler) ExportDataset(c *gin.Context) {
	datasetID := c.Param("id")
	format := c.DefaultQuery("format", "excel") // excel or csv

	// 获取数据预览(用于导出)
	limit := 10000 // 导出最多10000行
	result, err := h.datasetService.PreviewData(c.Request.Context(), datasetID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("dataset_%s", datasetID)
	var filepath string

	switch format {
	case "csv":
		filepath, err = h.exportService.ExportToCSV(c.Request.Context(), result.Data, filename)
	default:
		filepath, err = h.exportService.ExportToExcel(c.Request.Context(), result.Data, filename)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回文件下载
	c.FileAttachment(filepath, filepath[len(filepath)-len(filepath)+5:])
}

// ExportChartData 导出图表数据
func (h *ExportHandler) ExportChartData(c *gin.Context) {
	chartID := c.Param("id")
	format := c.DefaultQuery("format", "excel")

	// 获取图表数据
	data, err := h.chartService.GetChartData(c.Request.Context(), chartID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("chart_%s", chartID)
	var filePath string

	switch format {
	case "csv":
		filePath, err = h.exportService.ExportToCSV(c.Request.Context(), data, filename)
	default:
		filePath, err = h.exportService.ExportToExcel(c.Request.Context(), data, filename)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回文件下载
	c.FileAttachment(filePath, filepath.Base(filePath))
}
