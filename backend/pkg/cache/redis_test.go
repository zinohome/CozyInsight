package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRedisCache_SetGet 测试Set和Get操作
func TestRedisCache_SetGet(t *testing.T) {
	// 跳过集成测试(需要真实Redis)
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cache, err := NewRedisCache(&RedisConfig{
		Host: "localhost",
		Port: 6379,
		DB:   0,
	})
	if err != nil || cache == nil {
		t.Skip("Redis not available:", err)
		return
	}
	defer cache.Close()

	ctx := context.Background()

	// 测试字符串
	err = cache.Set(ctx, "test:string", "hello", 10*time.Second)
	assert.NoError(t, err)

	val, err := cache.Get(ctx, "test:string")
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)

	// 测试结构体
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	user := User{ID: 1, Name: "test"}
	err = cache.Set(ctx, "test:user", user, 10*time.Second)
	assert.NoError(t, err)

	val, err = cache.Get(ctx, "test:user")
	assert.NoError(t, err)
	assert.NotNil(t, val)

	// 清理
	cache.Delete(ctx, "test:string")
	cache.Delete(ctx, "test:user")
}

// TestRedisCache_Delete 测试删除操作
func TestRedisCache_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cache, err := NewRedisCache(&RedisConfig{
		Host: "localhost",
		Port: 6379,
		DB:   0,
	})
	assert.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// 设置值
	err = cache.Set(ctx, "test:delete", "value", 10*time.Second)
	assert.NoError(t, err)

	// 删除
	err = cache.Delete(ctx, "test:delete")
	assert.NoError(t, err)

	// 验证已删除
	_, err = cache.Get(ctx, "test:delete")
	assert.Error(t, err)
}

// TestRedisCache_Exists 测试键存在性检查
func TestRedisCache_Exists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cache, err := NewRedisCache(&RedisConfig{
		Host: "localhost",
		Port: 6379,
		DB:   0,
	})
	assert.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// 不存在的键
	exists, err := cache.Exists(ctx, "test:nonexist")
	assert.NoError(t, err)
	assert.False(t, exists)

	// 设置键
	err = cache.Set(ctx, "test:exist", "value", 10*time.Second)
	assert.NoError(t, err)

	// 检查存在
	exists, err = cache.Exists(ctx, "test:exist")
	assert.NoError(t, err)
	assert.True(t, exists)

	// 清理
	cache.Delete(ctx, "test:exist")
}
