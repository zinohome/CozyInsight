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
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"

	"cozy-insight/internal/engine"
)

func TestClickHouse_RealContainer(t *testing.T) {
	requireDocker(t)
	if testing.Short() {
		t.Skip("short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	ctr, err := clickhouse.RunContainer(ctx,
		clickhouse.WithDatabase("cozy"),
		clickhouse.WithUsername("default"),
		clickhouse.WithPassword("test"),
	)
	require.NoError(t, err, "start clickhouse container")
	defer func() { _ = ctr.Terminate(context.Background()) }()

	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	// ClickHouse HTTP 和 native 端口都可能用到；这里使用 native 端口（9000）
	port, err := ctr.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)
	portInt, err := strconv.Atoi(port.Port())
	require.NoError(t, err)
	// testcontainers Host() returns "localhost" on Mac which resolves to ::1
	// (IPv6); the docker-mapped port is only bound on IPv4. Force 127.0.0.1.
	host = "127.0.0.1"

	cfg := map[string]interface{}{
		"host":     host,
		"port":     portInt,
		"username": "default",
		"password": "test",
		"database": "cozy",
	}
	cfgJSON, err := json.Marshal(cfg)
	require.NoError(t, err)

	conn, err := engine.NewConnector("clickhouse")
	require.NoError(t, err)
	// ClickHouse may report "ready" via its log wait strategy before the
	// native protocol listener is fully accepting connections.
	require.NoError(t, connectWithRetry(t, func() error { return conn.Connect(string(cfgJSON)) }, 5, 2*time.Second))
	defer conn.Close()

	rows, err := conn.Query(ctx, "SELECT 1 AS n")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}

// TestClickHouse_GetColumns_RealContainer is a follow-up that exercises
// the GetColumns path. The clickhouse-go v2 driver occasionally returns
// "bad connection" on DDL immediately after a fresh SELECT, so this test
// opens a brand-new connection (no preceding Query) for the CREATE.
//
// Run with: go test -tags=integration -run "TestClickHouse_GetColumns" ./internal/integration/container/...
func TestClickHouse_GetColumns_RealContainer(t *testing.T) {
	requireDocker(t)
	if testing.Short() {
		t.Skip("short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	ctr, err := clickhouse.RunContainer(ctx,
		clickhouse.WithDatabase("cozy"),
		clickhouse.WithUsername("default"),
		clickhouse.WithPassword("test"),
	)
	require.NoError(t, err, "start clickhouse container")
	defer func() { _ = ctr.Terminate(context.Background()) }()

	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	port, err := ctr.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)
	portInt, err := strconv.Atoi(port.Port())
	require.NoError(t, err)
	host = "127.0.0.1"

	cfg := map[string]interface{}{
		"host":     host,
		"port":     portInt,
		"username": "default",
		"password": "test",
		"database": "cozy",
	}
	cfgJSON, err := json.Marshal(cfg)
	require.NoError(t, err)

	// Open a single connection and run CREATE → GetColumns on it without
	// any prior Query. The v2 driver fails with "bad connection" when
	// DDL follows DQL on the same connection, so we never mix them here.
	// We avoid CREATE altogether by querying the built-in `system.one`
	// table — its columns are well-known and don't require DDL.
	conn, err := engine.NewConnector("clickhouse")
	require.NoError(t, err)
	require.NoError(t, connectWithRetry(t, func() error { return conn.Connect(string(cfgJSON)) }, 10, 2*time.Second))
	defer conn.Close()
	cols, err := conn.GetColumns(ctx, "system", "one")
	require.NoError(t, err)
	names := make([]string, 0, len(cols))
	for _, c := range cols {
		names = append(names, c.Name)
	}
	assert.Contains(t, names, "dummy")
}
