package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"cozy-insight/internal/service"
)

type ExportHandler struct {
	chartService *service.ChartService
}

func NewExportHandler(chartService *service.ChartService) *ExportHandler {
	return &ExportHandler{chartService: chartService}
}

func (h *ExportHandler) ExportCSV(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	resp, err := h.chartService.GetData(c.Request.Context(), id, nil, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=chart-%d.csv", id))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Header
	var headers []string
	for _, d := range resp.Dimensions {
		headers = append(headers, d)
	}
	for _, m := range resp.Metrics {
		headers = append(headers, m)
	}
	writer.Write(headers)

	// Rows
	for _, row := range resp.Data {
		var record []string
		for _, d := range resp.Dimensions {
			record = append(record, fmt.Sprintf("%v", row[d]))
		}
		for _, m := range resp.Metrics {
			record = append(record, fmt.Sprintf("%v", row[m]))
		}
		writer.Write(record)
	}
}

func (h *ExportHandler) ExportExcel(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	resp, err := h.chartService.GetData(c.Request.Context(), id, nil, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheetName := "数据"
	f.SetSheetName("Sheet1", sheetName)

	// Header
	var headers []string
	for _, d := range resp.Dimensions {
		headers = append(headers, d)
	}
	for _, m := range resp.Metrics {
		headers = append(headers, m)
	}

	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Rows
	for rowIdx, row := range resp.Data {
		colIdx := 0
		for _, d := range resp.Dimensions {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, row[d])
			colIdx++
		}
		for _, m := range resp.Metrics {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, row[m])
			colIdx++
		}
	}

	// Auto-filter
	lastCol, _ := excelize.CoordinatesToCellName(len(headers), 1)
	f.AutoFilter(sheetName, "A1", lastCol, [])

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=chart-%d.xlsx", id))

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
	}
}
