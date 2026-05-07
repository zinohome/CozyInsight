package engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCacheService 模拟缓存服务
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func TestNewCalciteClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *CalciteConfig
		wantErr bool
	}{
		{
			name:    "nil config should fail",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid config",
			config: &CalciteConfig{
				AvaticaURL:      "http://localhost:8765/",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 1 * time.Hour,
			},
			wantErr: false, // 如果Avatica服务未运行，此测试会失败
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &MockCacheService{}
			client, err := NewCalciteClient(tt.config, cache)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				// 注意：此测试需要Avatica Server运行
				if err != nil {
					t.Skipf("Avatica server not available: %v", err)
				}
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if client != nil {
					defer client.Close()
				}
			}
		})
	}
}

func TestCalciteClient_GenerateCacheKey(t *testing.T) {
	client := &CalciteClient{}

	tests := []struct {
		name   string
		sql    string
		params []interface{}
		want   string
	}{
		{
			name:   "simple query",
			sql:    "SELECT 1",
			params: nil,
		},
		{
			name:   "query with params",
			sql:    "SELECT * FROM users WHERE id = ?",
			params: []interface{}{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := client.generateCacheKey(tt.sql, tt.params...)
			key2 := client.generateCacheKey(tt.sql, tt.params...)

			// 相同的输入应该产生相同的缓存键
			assert.Equal(t, key1, key2)
			assert.Contains(t, key1, "query:")
		})
	}
}

func TestCalciteClient_ExecuteQuery_WithCache(t *testing.T) {
	// 此测试需要运行中的Avatica Server
	t.Skip("Integration test - requires Avatica Server")

	cfg := &CalciteConfig{
		AvaticaURL:      "http://localhost:8765/",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 1 * time.Hour,
	}

	cache := &MockCacheService{}
	// 模拟缓存未命中
	cache.On("Get", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	client, err := NewCalciteClient(cfg, cache)
	assert.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	result, err := client.ExecuteQuery(ctx, "SELECT 1 as num")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// 验证缓存被调用
	cache.AssertCalled(t, "Get", mock.Anything, mock.Anything)
	cache.AssertCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCalciteClient_Stats(t *testing.T) {
	t.Skip("Integration test - requires Avatica Server")

	cfg := &CalciteConfig{
		AvaticaURL:   "http://localhost:8765/",
		MaxOpenConns: 10,
	}

	client, err := NewCalciteClient(cfg, nil)
	assert.NoError(t, err)
	defer client.Close()

	stats := client.Stats()
	assert.Equal(t, 10, stats.MaxOpenConnections)
}
