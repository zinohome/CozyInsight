package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

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
