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
	"github.com/testcontainers/testcontainers-go/modules/mysql"

	"cozy-insight/internal/engine"
)

func TestMySQL_RealContainer(t *testing.T) {
	requireDocker(t)
	if testing.Short() {
		t.Skip("short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ctr, err := mysql.RunContainer(ctx,
		mysql.WithDatabase("cozy"),
		mysql.WithUsername("root"),
		mysql.WithPassword("test"),
	)
	require.NoError(t, err, "start mysql container")
	defer func() { _ = ctr.Terminate(context.Background()) }()

	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	port, err := ctr.MappedPort(ctx, "3306/tcp")
	require.NoError(t, err)
	portInt, err := strconv.Atoi(port.Port())
	require.NoError(t, err)
	// testcontainers Host() returns "localhost" on Mac which resolves to ::1
	// (IPv6); the docker-mapped port is only bound on IPv4. Force 127.0.0.1.
	host = "127.0.0.1"

	cfg := map[string]interface{}{
		"host":     host,
		"port":     portInt,
		"username": "root",
		"password": "test",
		"database": "cozy",
	}
	cfgJSON, err := json.Marshal(cfg)
	require.NoError(t, err)

	conn, err := engine.NewConnector("mysql")
	require.NoError(t, err)
	// MySQL reports "ready" via log message but may take a few seconds
	// before fully accepting auth.
	require.NoError(t, connectWithRetry(t, func() error { return conn.Connect(string(cfgJSON)) }, 5, 2*time.Second))
	defer conn.Close()

	// Query a built-in table to verify connectivity
	rows, err := conn.Query(ctx, "SELECT 1 AS n")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.EqualValues(t, 1, rows[0]["n"])
}
