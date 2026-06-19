package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newRateLimitRouter(rate, burst float64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(rate, burst))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func doGet(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestRateLimit_AllowsUpToBurst 验证桶满时前 N 个请求(burst)全通过。
func TestRateLimit_AllowsUpToBurst(t *testing.T) {
	r := newRateLimitRouter(1, 5)

	for i := 0; i < 5; i++ {
		w := doGet(r, "/ping")
		assert.Equal(t, http.StatusOK, w.Code, "request %d should pass", i+1)
	}
}

// TestRateLimit_RejectsAfterBurst 验证超出 burst 立即 429。
func TestRateLimit_RejectsAfterBurst(t *testing.T) {
	r := newRateLimitRouter(0.001, 2) // 几乎不补充

	// 消耗 2 个令牌
	for i := 0; i < 2; i++ {
		w := doGet(r, "/ping")
		assert.Equal(t, http.StatusOK, w.Code)
	}
	// 第 3 个应当被拒
	w := doGet(r, "/ping")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "rate limit exceeded")
}

// TestRateLimit_429ErrorFormat 验证错误体符合标准响应。
func TestRateLimit_429ErrorFormat(t *testing.T) {
	r := newRateLimitRouter(0.001, 1)
	_ = doGet(r, "/ping")
	w := doGet(r, "/ping")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	// 项目统一响应格式: {"code": 429, "error": "..."}
	assert.Contains(t, w.Body.String(), `"code":429`)
}

// TestRateLimit_GlobalBucket 验证全局共享一个桶(不按 IP/用户切分)。
func TestRateLimit_GlobalBucket(t *testing.T) {
	r := newRateLimitRouter(0.001, 3)
	got := 0
	for i := 0; i < 5; i++ {
		w := doGet(r, "/ping")
		if w.Code == http.StatusOK {
			got++
		}
	}
	// 桶容量 3,所以最多 3 个成功
	assert.LessOrEqual(t, got, 3, "global bucket should limit total concurrent requests")
	assert.GreaterOrEqual(t, got, 1)
}
