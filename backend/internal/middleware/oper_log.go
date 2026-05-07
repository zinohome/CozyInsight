package middleware

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func OperationLog(repo *repository.OperationLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		w := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		duration := time.Since(start).Milliseconds()

		path := c.Request.URL.Path
		if path == "/health" || path == "/api/v1/auth/register" || path == "/api/v1/auth/login" {
			return
		}

		userID := GetUserID(c)
		username := GetUsername(c)

		body := string(bodyBytes)
		if len(body) > 4096 {
			body = body[:4096] + "..."
		}

		query := c.Request.URL.RawQuery
		if len(query) > 1024 {
			query = query[:1024] + "..."
		}

		log := &model.OperationLog{
			UserID:     userID,
			Username:   username,
			Method:     c.Request.Method,
			Path:       path,
			Query:      query,
			Body:       body,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			StatusCode: c.Writer.Status(),
			Duration:   duration,
		}

		if len(c.Errors) > 0 {
			log.ErrorMessage = c.Errors.Last().Error()
		}

		go repo.Create(context.Background(), log)
	}
}
