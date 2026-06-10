package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPoolConnector struct{}

func (m *mockPoolConnector) Connect(string) error                                  { return nil }
func (m *mockPoolConnector) Close() error                                          { return nil }
func (m *mockPoolConnector) Query(context.Context, string, ...interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockPoolConnector) GetColumns(context.Context, string, string) ([]ColumnInfo, error) {
	return nil, nil
}

func TestConnectorPool_Get_CreatesNew(t *testing.T) {
	pool := NewConnectorPool()
	require.NotNil(t, pool)
	require.NotNil(t, pool.pools)
	assert.Len(t, pool.pools, 0)
}

func TestConnectorPool_Get_ReusesExisting(t *testing.T) {
	pool := NewConnectorPool()
	mock := &mockPoolConnector{}
	pool.pools[1] = mock

	conn, err := pool.Get(1, "mysql", `{}`)
	require.NoError(t, err)
	assert.Equal(t, mock, conn)
	assert.Len(t, pool.pools, 1)
}

func TestConnectorPool_Remove(t *testing.T) {
	pool := NewConnectorPool()
	mock := &mockPoolConnector{}
	pool.pools[1] = mock

	pool.Remove(1)
	assert.Len(t, pool.pools, 0)
}

func TestConnectorPool_Remove_NotFound(t *testing.T) {
	pool := NewConnectorPool()
	pool.Remove(999) // should not panic
	assert.Len(t, pool.pools, 0)
}

func TestConnectorPool_Close(t *testing.T) {
	pool := NewConnectorPool()
	pool.pools[1] = &mockPoolConnector{}
	pool.pools[2] = &mockPoolConnector{}

	pool.Close()
	assert.Len(t, pool.pools, 0)
}

func TestConnectorPool_Register(t *testing.T) {
	pool := NewConnectorPool()
	mock := &mockPoolConnector{}
	pool.Register(1, mock)
	assert.Len(t, pool.pools, 1)
	assert.Equal(t, mock, pool.pools[1])
}

func TestConnectorPool_Register_NilPool(t *testing.T) {
	// 触发 pools==nil 时的 lazy init 分支
	pool := &ConnectorPool{}
	pool.Register(1, &mockPoolConnector{})
	assert.NotNil(t, pool.pools)
	assert.Len(t, pool.pools, 1)
}

func TestConnectorPool_Get_UnsupportedType(t *testing.T) {
	pool := NewConnectorPool()
	_, err := pool.Get(1, "unsupported-type", "{}")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestConnectorPool_Get_ConnectError(t *testing.T) {
	pool := NewConnectorPool()
	// api 类型 Connect 需要 url 字段；空配置触发 Connect 失败路径
	_, err := pool.Get(1, "api", `{}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connect")
}
