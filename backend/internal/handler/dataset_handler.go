package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

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

func (h *DatasetHandler) Preview(c *gin.Context) {
	id := c.Param("id")
	data, err := h.svc.PreviewData(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
