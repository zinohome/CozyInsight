package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/internal/dto"
	"cozy-insight/pkg/cache"
)

func TestCacheService_ChartData(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := cache.NewRedisClient(mr.Addr())
	svc := NewCacheService(redisClient)
	ctx := context.Background()
	data := &dto.ChartDataResponse{
		Dimensions: []string{"month"},
		Metrics:    []string{"sales"},
		Data:       []map[string]interface{}{{"month": "Jan", "sales": 100}},
	}
	err := svc.SetChartData(ctx, 1, `{"dimensions":[]}`, data, time.Minute)
	require.NoError(t, err)
	cached, err := svc.GetChartData(ctx, 1, `{"dimensions":[]}`)
	require.NoError(t, err)
	assert.Equal(t, data.Dimensions, cached.Dimensions)
	assert.Equal(t, data.Metrics, cached.Metrics)
	assert.Len(t, cached.Data, 1)
}

func TestCacheService_ChartData_Miss(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := cache.NewRedisClient(mr.Addr())
	svc := NewCacheService(redisClient)
	ctx := context.Background()
	_, err := svc.GetChartData(ctx, 1, `{"dimensions":[]}`)
	assert.Error(t, err)
}
