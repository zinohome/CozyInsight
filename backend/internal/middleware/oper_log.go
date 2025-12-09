package middleware

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OperLogMiddleware 操作日志中间件
func OperLogMiddleware(operLogRepo repository.OperLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 跳过特定路径
		if c.Request.URL.Path == "/health" || 
		   c.Request.URL.Path == "/api/v1/auth/login" {
			return
		}

		// 构造操作日志
		userID, exists := c.Get("userID")
		if !exists {
			return // 未认证请求不记录
		}

		username, _ := c.Get("username")

		log := &model.SysOperLog{
			ID:         uuid.New().String(),
			UserID:     userID.(string),
			Username:   username.(string),
			Module:     extractModule(c.Request.URL.Path),
			Action:     mapMethodToAction(c.Request.Method),
			ResourceID: c.Param("id"),
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Status:     1,
			CreateTime: startTime.UnixMilli(),
		}

		// 记录请求详情
		detail := map[string]interface{}{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"duration": time.Since(startTime).Milliseconds(),
		}
		
		if detailJSON, err := json.Marshal(detail); err == nil {
			log.Detail = string(detailJSON)
		}

		// 异步保存日志
		go func() {
			operLogRepo.Create(c.Request.Context(), log)
		}()
	}
}

// extractModule 从路径提取模块名
func extractModule(path string) string {
	// /api/v1/datasource/xxx -> datasource
	parts := splitPath(path)
	if len(parts) >= 3 {
		return parts[2]
	}
	return "unknown"
}

// mapMethodToAction 映射HTTP方法到操作类型
func mapMethodToAction(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	case "GET":
		return "view"
	default:
		return "unknown"
	}
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, ch := range path {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
