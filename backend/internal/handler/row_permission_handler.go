package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/service"
)

type RowPermissionHandler struct {
	service *service.RowPermissionService
}

func NewRowPermissionHandler(service *service.RowPermissionService) *RowPermissionHandler {
	return &RowPermissionHandler{service: service}
}

func (h *RowPermissionHandler) List(c *gin.Context) {
	datasetID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	list, err := h.service.ListByDataset(c.Request.Context(), datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *RowPermissionHandler) Create(c *gin.Context) {
	datasetID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		FieldName string `json:"fieldName" binding:"required"`
		Operator  string `json:"operator" binding:"required"`
		Value     string `json:"value"`
		UserAttr  string `json:"userAttr" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	rp, err := h.service.Create(c.Request.Context(), datasetID, req.FieldName, req.Operator, req.Value, req.UserAttr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": rp})
}

func (h *RowPermissionHandler) Delete(c *gin.Context) {
	permID, ok := parseUintParam(c, "permId")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), permID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}
