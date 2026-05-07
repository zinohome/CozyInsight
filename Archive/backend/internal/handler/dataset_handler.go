package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DatasetHandler struct {
	svc service.DatasetService
}

func NewDatasetHandler(svc service.DatasetService) *DatasetHandler {
	return &DatasetHandler{svc: svc}
}

func (h *DatasetHandler) CreateGroup(c *gin.Context) {
	var group model.DatasetGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.CreateGroup(c.Request.Context(), &group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *DatasetHandler) ListGroups(c *gin.Context) {
	list, err := h.svc.ListGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *DatasetHandler) ListTables(c *gin.Context) {
	list, err := h.svc.ListTables(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *DatasetHandler) CreateTable(c *gin.Context) {
	var table model.DatasetTable
	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.CreateTable(c.Request.Context(), &table); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, table)
}

// Preview 预览数据集数据
func (h *DatasetHandler) Preview(c *gin.Context) {
	id := c.Param("id")

	// 获取 limit 参数，默认100
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	result, err := h.svc.PreviewData(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetFields 获取数据集字段信息
func (h *DatasetHandler) GetFields(c *gin.Context) {
	id := c.Param("id")

	fields, err := h.svc.GetFields(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"fields": fields})
}

// SyncFields 同步字段到数据库
func (h *DatasetHandler) SyncFields(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.SyncFields(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "字段同步成功"})
}
