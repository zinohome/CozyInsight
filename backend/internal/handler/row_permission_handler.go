package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RowPermissionHandler struct {
	service service.RowPermissionService
}

func NewRowPermissionHandler(service service.RowPermissionService) *RowPermissionHandler {
	return &RowPermissionHandler{service: service}
}

// Create 创建行级权限
func (h *RowPermissionHandler) Create(c *gin.Context) {
	var perm model.DatasetRowPermissions

	if err := c.ShouldBindJSON(&perm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	perm.CreateBy = userID.(string)

	if err := h.service.Create(c.Request.Context(), &perm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, perm)
}

// List 获取数据集的行级权限列表
func (h *RowPermissionHandler) List(c *gin.Context) {
	datasetID := c.Param("datasetId")

	perms, err := h.service.ListByDataset(c.Request.Context(), datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// Delete 删除行级权限
func (h *RowPermissionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
