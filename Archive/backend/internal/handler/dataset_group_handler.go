package handler

import (
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DatasetGroupHandler struct {
	service service.DatasetGroupService
}

func NewDatasetGroupHandler(service service.DatasetGroupService) *DatasetGroupHandler {
	return &DatasetGroupHandler{service: service}
}

// GetTree 获取数据集分组树
func (h *DatasetGroupHandler) GetTree(c *gin.Context) {
	tree, err := h.service.GetGroupTree(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tree)
}

// List 获取所有分组
func (h *DatasetGroupHandler) List(c *gin.Context) {
	groups, err := h.service.ListGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}
