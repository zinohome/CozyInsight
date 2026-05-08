package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisClient_SetAndGet(t *testing.T) {
	mr := miniredis.RunT(t)
	client := NewRedisClient(mr.Addr())
	ctx := context.Background()
	err := client.Set(ctx, "test-key", "test-value", time.Minute)
	require.NoError(t, err)
	val, err := client.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, "test-value", val)
}

func TestRedisClient_KeyNotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	client := NewRedisClient(mr.Addr())
	ctx := context.Background()
	_, err := client.Get(ctx, "missing-key")
	assert.Error(t, err)
	assert.True(t, client.IsNotFound(err))
}

func TestRedisClient_TTL(t *testing.T) {
	mr := miniredis.RunT(t)
	client := NewRedisClient(mr.Addr())
	ctx := context.Background()
	err := client.Set(ctx, "ttl-key", "val", time.Second)
	require.NoError(t, err)
	exists, err := client.Exists(ctx, "ttl-key")
	require.NoError(t, err)
	assert.True(t, exists)
	mr.FastForward(time.Second * 2)
	exists, err = client.Exists(ctx, "ttl-key")
	require.NoError(t, err)
	assert.False(t, exists)
}
