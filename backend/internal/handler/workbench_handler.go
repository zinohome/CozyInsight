package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/service"
)

type WorkbenchHandler struct {
	service *service.WorkbenchService
}

func NewWorkbenchHandler(service *service.WorkbenchService) *WorkbenchHandler {
	return &WorkbenchHandler{service: service}
}

func (h *WorkbenchHandler) GetStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stats, err := h.service.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

func (h *WorkbenchHandler) GetRecentViews(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.service.ListRecentViews(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *WorkbenchHandler) RecordVisit(c *gin.Context) {
	var req dto.RecordVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.service.RecordVisit(c.Request.Context(), userID, req.ResourceType, req.ResourceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200})
}

func (h *WorkbenchHandler) GetFavorites(c *gin.Context) {
	userID := middleware.GetUserID(c)
	list, err := h.service.ListFavorites(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *WorkbenchHandler) AddFavorite(c *gin.Context) {
	var req dto.AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.service.AddFavorite(c.Request.Context(), userID, req.ResourceType, req.ResourceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200})
}

func (h *WorkbenchHandler) DeleteFavorite(c *gin.Context) {
	resourceType := c.Param("type")
	resourceID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.service.DeleteFavorite(c.Request.Context(), userID, resourceType, resourceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200})
}
