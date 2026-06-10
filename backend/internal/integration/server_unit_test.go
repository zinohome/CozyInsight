package integration

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTestServer_Smoke 验证 TestServer 能起能停。
func TestNewTestServer_Smoke(t *testing.T) {
	ts := NewTestServer(t, nil)
	require.NotNil(t, ts)
	require.NotEmpty(t, ts.URL)
	ts.Close()
}

func TestDefaultTestConfig(t *testing.T) {
	cfg := DefaultTestConfig()
	assert.Equal(t, "test", cfg.Server.Mode)
	assert.Equal(t, "test-secret-key-do-not-use-in-prod", cfg.JWT.Secret)
	assert.Equal(t, 24*time.Hour, cfg.JWT.ExpireHours)
}

// TestLoginAs_BadPassword 验证未注册用户时登录返回非 200。
// 配合 sqlmock 未注册任何 expectation，期望数据库驱动层报错。
func TestLoginAs_DBError(t *testing.T) {
	ts := NewTestServer(t, nil)
	// 不设置任何 sqlmock expectation → 第一次查询会返回 "call to QueryContext was not expected"
	body := `{"username":"alice","password":"x"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/login", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	// 数据库错误 → handler 应当返回 500 或者 200+业务错误码
	assert.NotEqual(t, http.StatusOK, resp.StatusCode,
		"login should fail when DB returns error")
	assert.NoError(t, ts.Mock.ExpectationsWereMet())
}
