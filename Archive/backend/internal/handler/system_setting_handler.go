package handler

import (
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemSettingHandler struct {
	service service.SystemSettingService
}

func NewSystemSettingHandler(service service.SystemSettingService) *SystemSettingHandler {
	return &SystemSettingHandler{service: service}
}

// Get 获取配置
func (h *SystemSettingHandler) Get(c *gin.Context) {
	key := c.Param("key")

	setting, err := h.service.Get(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// Set 设置配置
func (h *SystemSettingHandler) Set(c *gin.Context) {
	var req struct {
		Type  string `json:"type" binding:"required"`
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	err := h.service.Set(c.Request.Context(), req.Type, req.Key, req.Value, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// ListByType 按类型获取配置列表
func (h *SystemSettingHandler) ListByType(c *gin.Context) {
	settingType := c.Param("type")

	settings, err := h.service.GetByType(c.Request.Context(), settingType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// Delete 删除配置
func (h *SystemSettingHandler) Delete(c *gin.Context) {
	key := c.Param("key")

	err := h.service.Delete(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
