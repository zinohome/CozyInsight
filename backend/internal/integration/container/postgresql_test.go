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
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"cozy-insight/internal/engine"
)

func TestPostgres_RealContainer(t *testing.T) {
	requireDocker(t)
	if testing.Short() {
		t.Skip("short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ctr, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("cozy"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("test"),
	)
	require.NoError(t, err, "start postgres container")
	defer func() { _ = ctr.Terminate(context.Background()) }()

	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	port, err := ctr.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	portInt, err := strconv.Atoi(port.Port())
	require.NoError(t, err)
	// testcontainers Host() returns "localhost" on Mac which resolves to ::1
	// (IPv6); the docker-mapped port is only bound on IPv4. Force 127.0.0.1.
	host = "127.0.0.1"

	cfg := map[string]interface{}{
		"host":     host,
		"port":     portInt,
		"username": "postgres",
		"password": "test",
		"database": "cozy",
		"sslmode":  "disable",
	}
	cfgJSON, err := json.Marshal(cfg)
	require.NoError(t, err)

	conn, err := engine.NewConnector("postgresql")
	require.NoError(t, err)
	// Postgres container reports "ready" before fully accepting auth; retry a few times.
	require.NoError(t, connectWithRetry(t, func() error { return conn.Connect(string(cfgJSON)) }, 5, 2*time.Second))
	defer conn.Close()

	rows, err := conn.Query(ctx, "SELECT 1 AS n")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.EqualValues(t, 1, rows[0]["n"])

	// Create a temp table and verify GetColumns reads INFORMATION_SCHEMA.
	_, err = conn.Query(ctx, `CREATE TABLE IF NOT EXISTS public.t1 (id SERIAL PRIMARY KEY, name TEXT, score NUMERIC(10,2))`)
	require.NoError(t, err)
	cols, err := conn.GetColumns(ctx, "public", "t1")
	require.NoError(t, err)
	names := make([]string, 0, len(cols))
	for _, c := range cols {
		names = append(names, c.Name)
	}
	assert.Contains(t, names, "id")
	assert.Contains(t, names, "name")
	assert.Contains(t, names, "score")
}
