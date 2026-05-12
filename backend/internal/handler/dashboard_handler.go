package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/service"
)

type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler(service *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Create(c *gin.Context) {
	var req dto.CreateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)

	d, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": d})
}

func (h *DashboardHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	d, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": d})
}

func (h *DashboardHandler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *DashboardHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.service.Update(c.Request.Context(), id, &req, userID); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DashboardHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.service.Delete(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DashboardHandler) EnableShare(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		Password    string `json:"password"`
		ExpireHours int    `json:"expireHours"`
	}
	_ = c.ShouldBindJSON(&req)
	userID := middleware.GetUserID(c)
	token, err := h.service.EnableShare(c.Request.Context(), id, userID, req.Password, req.ExpireHours)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": token})
}

func (h *DashboardHandler) DisableShare(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.service.DisableShare(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DashboardHandler) AddChart(c *gin.Context) {
	dashboardID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.AddChartToDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.AddChart(c.Request.Context(), dashboardID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DashboardHandler) GetCharts(c *gin.Context) {
	dashboardID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	list, err := h.service.GetCharts(c.Request.Context(), dashboardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *DashboardHandler) RemoveChart(c *gin.Context) {
	dashboardID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	chartID, ok := parseUintParam(c, "chartId")
	if !ok {
		return
	}
	if err := h.service.RemoveChart(c.Request.Context(), dashboardID, chartID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}
