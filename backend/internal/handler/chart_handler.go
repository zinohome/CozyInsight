package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChartHandler struct {
	svc          service.ChartService
	chartDataSvc service.ChartDataService
}

func NewChartHandler(svc service.ChartService, chartDataSvc service.ChartDataService) *ChartHandler {
	return &ChartHandler{
		svc:          svc,
		chartDataSvc: chartDataSvc,
	}
}

func (h *ChartHandler) Create(c *gin.Context) {
	var chart model.ChartView
	if err := c.ShouldBindJSON(&chart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Create(c.Request.Context(), &chart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chart)
}

func (h *ChartHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var chart model.ChartView
	if err := c.ShouldBindJSON(&chart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	chart.ID = id

	if err := h.svc.Update(c.Request.Context(), &chart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chart)
}

func (h *ChartHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ChartHandler) Get(c *gin.Context) {
	id := c.Param("id")
	chart, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chart)
}

func (h *ChartHandler) List(c *gin.Context) {
	sceneId := c.Query("sceneId")
	list, err := h.svc.List(c.Request.Context(), sceneId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

// GetData 获取图表数据
func (h *ChartHandler) GetData(c *gin.Context) {
	id := c.Param("id")
	data, err := h.chartDataSvc.GetChartData(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
