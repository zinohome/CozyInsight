//go:build integration
// +build integration

package container

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"cozy-insight/internal/engine"
)

// TestOracle_RealContainer 真实 Oracle 容器集成测试。
//
// gvenzl/oracle-free 镜像首次启动较慢（>60s），所以这个测试放在 -short 模式
// 之外，并使用 5 分钟的 timeout。
//
// v0.42.0 的 testcontainers-go 还没有专用的 modules/oracle 包，所以用
// GenericContainer 启动 gvenzl/oracle-free 镜像。
func TestOracle_RealContainer(t *testing.T) {
	requireDocker(t)
	if testing.Short() {
		t.Skip("short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "gvenzl/oracle-free:23-slim",
		ExposedPorts: []string{"1521/tcp"},
		Env: map[string]string{
			"ORACLE_PASSWORD":    "testpass",
			"APP_USER":           "testuser",
			"APP_USER_PASSWORD":  "testpass",
		},
		WaitingFor: wait.ForLog("DATABASE IS READY TO USE").
			WithStartupTimeout(4 * time.Minute),
	}
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "start oracle container")
	defer func() { _ = ctr.Terminate(context.Background()) }()

	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	port, err := ctr.MappedPort(ctx, "1521/tcp")
	require.NoError(t, err)
	portInt, err := strconv.Atoi(port.Port())
	require.NoError(t, err)
	// Force 127.0.0.1 on Mac (IPv6 resolution race).
	host = "127.0.0.1"

	cfg := map[string]interface{}{
		"host":     host,
		"port":     portInt,
		"username": "testuser",
		"password": "testpass",
		"database": "FREEPDB1",
	}
	cfgJSON, err := json.Marshal(cfg)
	require.NoError(t, err)

	conn, err := engine.NewConnector("oracle")
	require.NoError(t, err)
	// Oracle 容器在 listener 起来后还要初始化 pluggable DB；多等几次。
	require.NoError(t, connectWithRetry(t, func() error { return conn.Connect(string(cfgJSON)) }, 10, 3*time.Second))
	defer conn.Close()

	// 查 DUAL 表验证连通性（DUAL 在每个 Oracle 实例中自动存在）
	rows, err := conn.Query(ctx, "SELECT 1 AS n FROM DUAL")
	require.NoError(t, err)
	assert.Len(t, rows, 1)

	// Create a temp table and verify GetColumns reads ALL_TAB_COLUMNS.
	// Oracle identifiers are upper-case by default; use unquoted names.
	_, err = conn.Query(ctx, `CREATE TABLE t1 (id NUMBER(10) PRIMARY KEY, name VARCHAR2(64), score NUMBER(10,2))`)
	require.NoError(t, err)
	cols, err := conn.GetColumns(ctx, "TESTUSER", "T1")
	require.NoError(t, err)
	names := make([]string, 0, len(cols))
	for _, c := range cols {
		names = append(names, c.Name)
	}
	assert.Contains(t, names, "ID")
	assert.Contains(t, names, "NAME")
	assert.Contains(t, names, "SCORE")
}
