package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"cozy-insight/internal/dto"
	"cozy-insight/pkg/cache"
)

type CacheService struct {
	redis *cache.RedisClient
}

func NewCacheService(redis *cache.RedisClient) *CacheService {
	return &CacheService{redis: redis}
}

func (s *CacheService) chartCacheKey(chartID uint64, config string) string {
	hash := sha256.Sum256([]byte(config))
	return fmt.Sprintf("chart:data:%d:%x", chartID, hash[:8])
}

func (s *CacheService) GetChartData(ctx context.Context, chartID uint64, config string) (*dto.ChartDataResponse, error) {
	key := s.chartCacheKey(chartID, config)
	val, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var resp dto.ChartDataResponse
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return nil, fmt.Errorf("cache unmarshal failed: %w", err)
	}
	return &resp, nil
}

func (s *CacheService) SetChartData(ctx context.Context, chartID uint64, config string, data *dto.ChartDataResponse, ttl time.Duration) error {
	key := s.chartCacheKey(chartID, config)
	val, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("cache marshal failed: %w", err)
	}
	return s.redis.Set(ctx, key, string(val), ttl)
}

func (s *CacheService) InvalidateChartData(ctx context.Context, chartID uint64) error {
	pattern := fmt.Sprintf("chart:data:%d:*", chartID)
	keys, err := s.redis.ScanKeys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}
	if err := s.redis.DelKeys(ctx, keys...); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}
