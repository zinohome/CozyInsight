//go:build integration
// +build integration

package container

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/engine"
)

func TestSQLite_RealContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := `{"database":":memory:"}`
	conn, err := engine.NewConnector("sqlite")
	require.NoError(t, err)
	require.NoError(t, conn.Connect(cfg))
	defer conn.Close()

	_, err = conn.Query(ctx, `CREATE TABLE t1 (id INTEGER PRIMARY KEY, name TEXT, score REAL)`)
	require.NoError(t, err)
	_, err = conn.Query(ctx, `INSERT INTO t1 (id, name, score) VALUES (1, 'alice', 95.5), (2, 'bob', 88.0)`)
	require.NoError(t, err)

	rows, err := conn.Query(ctx, `SELECT id, name, score FROM t1 ORDER BY id`)
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, int64(1), rows[0]["id"])
	assert.Equal(t, "alice", rows[0]["name"])

	cols, err := conn.GetColumns(ctx, "", "t1")
	require.NoError(t, err)
	names := make([]string, 0, len(cols))
	for _, c := range cols {
		names = append(names, c.Name)
	}
	assert.Contains(t, names, "id")
	assert.Contains(t, names, "name")
	assert.Contains(t, names, "score")
}

// TestSQLite_QueryError 触发 sqlite Query 失败路径
func TestSQLite_QueryError(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	conn, err := engine.NewConnector("sqlite")
	require.NoError(t, err)
	require.NoError(t, conn.Connect(`{"database":":memory:"}`))
	defer conn.Close()

	// 无效 SQL 触发 query failed 错误
	_, err = conn.Query(context.Background(), "SELECT * FROM nonexistent_table")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
}

// TestSQLite_GetColumnsError 触发 GetColumns 失败路径（关闭连接后调用）
func TestSQLite_GetColumnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	conn, err := engine.NewConnector("sqlite")
	require.NoError(t, err)
	require.NoError(t, conn.Connect(`{"database":":memory:"}`))
	require.NoError(t, conn.Close())

	_, err = conn.GetColumns(context.Background(), "", "t1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}
