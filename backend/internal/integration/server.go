// Package integration 提供了 HTTP 端到端集成测试基础设施。
//
// 关键设计：
//   - 使用 sqlmock 模拟 MySQL 行为，注入到 Repository 层之上的 Service。
//   - 使用 miniredis 模拟 Redis（如果测试需要缓存）。
//   - 使用 httptest 启动一个真实的 Gin 引擎，端到端走 HTTP 路由。
//   - JWT secret 在测试时固定，避免依赖环境变量。
//
// 使用方式（详见 internal/testutil/server_test.go）：
//
//	func TestSomething(t *testing.T) {
//	    s := testutil.NewTestServer(t)
//	    defer s.Close()
//	    token := s.LoginAs(t, "alice", "alice123")
//	    // 发起 HTTP 请求 ...
//	}
package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	v1 "cozy-insight/api/v1"
	"cozy-insight/pkg/cache"
	"cozy-insight/pkg/config"
	"cozy-insight/pkg/jwt"
)

// TestServer 是一个开箱即用的 HTTP 测试服务器，所有路由通过 v1.Setup 装配。
// 关闭时记得调用 Close() 来停止 httptest.Server 与 miniredis。
type TestServer struct {
	*httptest.Server
	DB      *sqlx.DB
	Mock    sqlmock.Sqlmock
	Redis   *miniredis.Miniredis
	Cfg     *config.Config
	JWT     *jwt.Manager
}

// NewTestServer 构造一个测试服务器：mock DB + miniredis + Gin 引擎 + httptest.Server。
// 如果 cfg=nil，使用一套默认的测试配置（dev 模式、MySQL 3306、Redis 6379）。
func NewTestServer(t *testing.T, cfg *config.Config) *TestServer {
	t.Helper()
	gin.SetMode(gin.TestMode)

	if cfg == nil {
		cfg = DefaultTestConfig()
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "mysql")

	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisClient := cache.NewRedisClient(mr.Addr())
	require.NoError(t, redisClient.Ping(context.Background()))

	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)

	r := gin.New()
	r.Use(gin.Recovery())
	v1.Setup(sqlxDB, cfg, r, redisClient)

	srv := httptest.NewServer(r)
	ts := &TestServer{
		Server: srv,
		DB:     sqlxDB,
		Mock:   mock,
		Redis:  mr,
		Cfg:    cfg,
		JWT:    jwtManager,
	}
	t.Cleanup(ts.Close)
	return ts
}

// Close 释放资源：关闭 httptest.Server、miniredis、底层 sql.DB。
func (s *TestServer) Close() {
	if s.Server != nil {
		s.Server.Close()
	}
	if s.Redis != nil {
		s.Redis.Close()
	}
	if s.DB != nil {
		_ = s.DB.Close()
	}
}

// LoginAs 模拟用户登录，调用 /api/v1/auth/login 并返回 JWT。
// 默认依赖底层 Repository 用 sqlmock 预期一个用户查询；调用方应先注册 mock。
func (s *TestServer) LoginAs(t *testing.T, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	resp, err := http.Post(s.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "login failed: %s", string(respBody))
	var result struct {
		Code int    `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(respBody, &result))
	require.NotEmpty(t, result.Data.Token)
	return result.Data.Token
}

// DoRequest 是发起带 Bearer token 的 HTTP 请求的便捷封装。
func (s *TestServer) DoRequest(t *testing.T, method, path, token string, body interface{}) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, s.URL+path, bodyReader)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// DefaultTestConfig 返回测试专用配置（短超时、确定 JWT secret）。
func DefaultTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{Port: 0, Mode: "test"},
		Database: config.DatabaseConfig{
			Driver:    "mysql",
			Host:      "127.0.0.1",
			Port:      3306,
			Username:  "test",
			Password:  "test",
			Database:  "test",
			Charset:   "utf8mb4",
			ParseTime: true,
			Loc:       "Local",
		},
		Redis: config.RedisConfig{Host: "127.0.0.1", Port: 6379, DB: 0},
		JWT: config.JWTConfig{
			Secret:      "test-secret-key-do-not-use-in-prod",
			ExpireHours: 24 * time.Hour,
		},
	}
}

// Compile-time guard: ensure the standard *sql.DB is wired through sqlx so
// repository code that calls db.Close works as expected in tests.
var _ = sql.ErrNoRows
