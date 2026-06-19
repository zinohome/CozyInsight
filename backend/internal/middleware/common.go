package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit 基于令牌桶的限流(全局共享,即所有请求共用一个桶)。
// rate:每秒补充的令牌数;burst:桶容量(=最大突发请求数)。
//
// 生产环境应改为按用户/IP 区分;此实现主要用于单实例演示。
// 触发限流时返回 429 + 标准错误格式。
func RateLimit(rate, burst float64) gin.HandlerFunc {
	bucket := &tokenBucket{
		tokens:   burst,
		capacity: burst,
		rate:     rate,
		lastTime: time.Now(),
	}
	return func(c *gin.Context) {
		if !bucket.allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":  429,
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// tokenBucket 简单内存令牌桶。
// 当前未做按用户/IP 切分;若多副本部署需替换为 Redis 实现。
type tokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	capacity float64
	rate     float64
	lastTime time.Time
}

// allow 尝试消费一个令牌,成功返回 true,桶空返回 false。
func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastTime).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastTime = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
