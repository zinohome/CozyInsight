package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"cozy-insight/pkg/jwt"
)

const ContextKeyUserID = "userID"
const ContextKeyUsername = "username"
const ContextKeyIsAdmin = "isAdmin"

func JWTAuth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := jwtManager.Parse(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyIsAdmin, claims.IsAdmin)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint64 {
	if v, ok := c.Get(ContextKeyUserID); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

func GetUsername(c *gin.Context) string {
	if v, ok := c.Get(ContextKeyUsername); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetIsAdmin(c *gin.Context) bool {
	if v, ok := c.Get(ContextKeyIsAdmin); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
