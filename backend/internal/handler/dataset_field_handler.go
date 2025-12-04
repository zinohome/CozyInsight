package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DatasetFieldHandler struct {
	svc service.DatasetFieldService
}

func NewDatasetFieldHandler(svc service.DatasetFieldService) *DatasetFieldHandler {
	return &DatasetFieldHandler{svc: svc}
}

// CreateField 创建字段
func (h *DatasetFieldHandler) CreateField(c *gin.Context) {
	var field model.DatasetTableField
	if err := c.ShouldBindJSON(&field); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.CreateField(c.Request.Context(), &field); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, field)
}

// UpdateField 更新字段
func (h *DatasetFieldHandler) UpdateField(c *gin.Context) {
	id := c.Param("fieldId")
	var field model.DatasetTableField
	if err := c.ShouldBindJSON(&field); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	field.ID = id

	if err := h.svc.UpdateField(c.Request.Context(), &field); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, field)
}

// DeleteField 删除字段
func (h *DatasetFieldHandler) DeleteField(c *gin.Context) {
	id := c.Param("fieldId")
	if err := h.svc.DeleteField(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetField 获取字段详情
func (h *DatasetFieldHandler) GetField(c *gin.Context) {
	id := c.Param("fieldId")
	field, err := h.svc.GetField(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, field)
}

// ListFields 获取字段列表
func (h *DatasetFieldHandler) ListFields(c *gin.Context) {
	tableId := c.Param("id")
	fields, err := h.svc.ListFields(c.Request.Context(), tableId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fields)
}

// SyncFields 同步字段
func (h *DatasetFieldHandler) SyncFields(c *gin.Context) {
	tableId := c.Param("id")
	if err := h.svc.SyncFieldsFromTable(c.Request.Context(), tableId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "字段同步成功"})
}
