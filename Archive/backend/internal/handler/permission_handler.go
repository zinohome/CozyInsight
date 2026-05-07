package handler

import (
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	service service.PermissionService
}

func NewPermissionHandler(service service.PermissionService) *PermissionHandler {
	return &PermissionHandler{service: service}
}

// List 获取权限列表
func (h *PermissionHandler) List(c *gin.Context) {
	permissions, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// GetRolePermissions 获取角色的权限
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("roleId")

	permissions, err := h.service.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// GrantToRole 授予权限给角色
func (h *PermissionHandler) GrantToRole(c *gin.Context) {
	roleID := c.Param("roleId")
	
	var req struct {
		PermissionIDs []string `json:"permissionIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.GrantPermissionToRole(c.Request.Context(), roleID, req.PermissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "granted"})
}

// RevokeFromRole 从角色撤销权限
func (h *PermissionHandler) RevokeFromRole(c *gin.Context) {
	roleID := c.Param("roleId")
	
	var req struct {
		PermissionIDs []string `json:"permissionIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.RevokePermissionFromRole(c.Request.Context(), roleID, req.PermissionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}

// GrantResourcePermission 授予资源权限
func (h *PermissionHandler) GrantResourcePermission(c *gin.Context) {
	var req struct {
		ResourceType string `json:"resourceType" binding:"required"`
		ResourceID   string `json:"resourceId" binding:"required"`
		TargetType   string `json:"targetType" binding:"required"`
		TargetID     string `json:"targetId" binding:"required"`
		Permission   string `json:"permission" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	
	if err := h.service.GrantResourcePermission(
		c.Request.Context(),
		req.ResourceType,
		req.ResourceID,
		req.TargetType,
		req.TargetID,
		req.Permission,
		userID.(string),
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "granted"})
}

// GetResourcePermissions 获取资源的权限配置
func (h *PermissionHandler) GetResourcePermissions(c *gin.Context) {
	resourceType := c.Query("resourceType")
	resourceID := c.Query("resourceId")

	if resourceType == "" || resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceType and resourceId are required"})
		return
	}

	permissions, err := h.service.GetResourcePermissions(c.Request.Context(), resourceType, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// CheckPermission 检查用户权限
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	userID, _ := c.Get("userID")
	resourceType := c.Query("resourceType")
	resourceID := c.Query("resourceId")
	action := c.Query("action")

	if resourceType == "" || resourceID == "" || action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceType, resourceId and action are required"})
		return
	}

	hasPermission, err := h.service.CheckPermission(
		c.Request.Context(),
		userID.(string),
		resourceType,
		resourceID,
		action,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hasPermission": hasPermission})
}
