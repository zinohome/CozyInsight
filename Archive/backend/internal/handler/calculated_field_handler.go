package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CalculatedFieldHandler struct {
	service service.CalculatedFieldService
}

func NewCalculatedFieldHandler(service service.CalculatedFieldService) *CalculatedFieldHandler {
	return &CalculatedFieldHandler{service: service}
}

// Create 创建计算字段
func (h *CalculatedFieldHandler) Create(c *gin.Context) {
	var req struct {
		DatasetTableID string `json:"datasetTableId" binding:"required"`
		FieldName      string `json:"fieldName" binding:"required"`
		DisplayName    string `json:"displayName"`
		Expression     string `json:"expression" binding:"required"`
		DataType       string `json:"dataType"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	field := &model.DatasetTableFieldCalculated{
		DatasetTableID: req.DatasetTableID,
		FieldName:      req.FieldName,
		DisplayName:    req.DisplayName,
		Expression:     req.Expression,
		DataType:       req.DataType,
	}

	if err := h.service.Create(c.Request.Context(), field); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, field)
}

// List 获取计算字段列表
func (h *CalculatedFieldHandler) List(c *gin.Context) {
	tableID := c.Param("tableId")

	fields, err := h.service.ListByTable(c.Request.Context(), tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fields)
}

// Delete 删除计算字段
func (h *CalculatedFieldHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
