package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	service service.RoleService
}

func NewRoleHandler(service service.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

// Create 创建角色
func (h *RoleHandler) Create(c *gin.Context) {
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文获取当前用户
	userID, _ := c.Get("userID")
	role.CreateBy = userID.(string)

	if err := h.service.Create(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// Update 更新角色
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	role.ID = id

	if err := h.service.Update(c.Request.Context(), &role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// Delete 删除角色
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Get 获取角色详情
func (h *RoleHandler) Get(c *gin.Context) {
	id := c.Param("id")

	role, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// List 获取角色列表
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}

// AssignToUser 分配角色给用户
func (h *RoleHandler) AssignToUser(c *gin.Context) {
	roleID := c.Param("id")
	
	var req struct {
		UserID string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AssignRoleToUser(c.Request.Context(), req.UserID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "assigned"})
}

// RemoveFromUser 从用户移除角色
func (h *RoleHandler) RemoveFromUser(c *gin.Context) {
	roleID := c.Param("id")
	userID := c.Query("userId")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	if err := h.service.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}

// GetUserRoles 获取用户的所有角色
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	roles, err := h.service.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}
