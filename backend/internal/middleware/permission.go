package middleware

import (
	"cozy-insight-backend/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(permissionService service.PermissionService, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// 从路径解析资源类型和ID
		resourceType, resourceID := parseResourceFromPath(c.Request.URL.Path)
		if resourceType == "" || resourceID == "" {
			// 如果无法解析资源,跳过权限检查(对于列表等操作)
			c.Next()
			return
		}

		// 检查权限
		hasPermission, err := permissionService.CheckPermission(
			c.Request.Context(),
			userID.(string),
			resourceType,
			resourceID,
			requiredPermission,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permission check failed"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleMiddleware 角色检查中间件
func RoleMiddleware(permissionService service.PermissionService, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		hasRole, err := permissionService.CheckUserHasRole(
			c.Request.Context(),
			userID.(string),
			requiredRole,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "role check failed"})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// parseResourceFromPath 从路径解析资源类型和ID
// 例如: /api/v1/datasource/123 -> ("datasource", "123")
func parseResourceFromPath(path string) (string, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	// 至少需要 api/v1/resource/id 四个部分
	if len(parts) < 4 {
		return "", ""
	}

	resourceType := parts[2] // api/v1/[resource]/id
	resourceID := parts[3]   // api/v1/resource/[id]

	// 验证资源类型
	validTypes := map[string]bool{
		"datasource": true,
		"dataset":    true,
		"chart":      true,
		"dashboard":  true,
	}

	if !validTypes[resourceType] {
		return "", ""
	}

	return resourceType, resourceID
}

// RequirePermission 资源权限装饰器辅助函数
func RequirePermission(action string) gin.HandlerFunc {
	// 这个函数需要在router中使用时注入permissionService
	// 实际使用时应该这样: RequirePermission(permissionService, "read")
	return func(c *gin.Context) {
		c.Set("requiredPermission", action)
		c.Next()
	}
}
