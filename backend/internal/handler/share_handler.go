package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShareHandler struct {
	service service.ShareService
}

func NewShareHandler(service service.ShareService) *ShareHandler {
	return &ShareHandler{service: service}
}

// Create 创建分享
func (h *ShareHandler) Create(c *gin.Context) {
	var share model.Share
	if err := c.ShouldBindJSON(&share); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	share.CreateBy = userID.(string)

	if err := h.service.CreateShare(c.Request.Context(), &share); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}

// Get 获取分享详情
func (h *ShareHandler) Get(c *gin.Context) {
	id := c.Param("id")

	share, err := h.service.GetShare(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}

// Delete 删除分享
func (h *ShareHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteShare(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// List 获取分享列表
func (h *ShareHandler) List(c *gin.Context) {
	resourceType := c.Query("resourceType")
	resourceID := c.Query("resourceId")

	shares, err := h.service.ListShares(c.Request.Context(), resourceType, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, shares)
}

// Validate 验证分享(公开访问)
func (h *ShareHandler) Validate(c *gin.Context) {
	token := c.Param("token")
	password := c.Query("password")

	share, err := h.service.ValidateShare(c.Request.Context(), token, password)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}
