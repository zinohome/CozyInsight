package engine

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/apache/calcite-avatica-go/v5"
)

// CalciteClient Avatica客户端，用于执行SQL查询
type CalciteClient struct {
	db    *sql.DB
	cache CacheService
}

// CalciteConfig Avatica配置
type CalciteConfig struct {
	AvaticaURL      string        `mapstructure:"avatica_url"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// CacheService 缓存服务接口
type CacheService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// NewCalciteClient 创建新的Calcite客户端
func NewCalciteClient(cfg *CalciteConfig, cache CacheService) (*CalciteClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("calcite config is required")
	}

	// 连接Avatica Server
	// 连接字符串格式: http://host:port/
	db, err := sql.Open("avatica", cfg.AvaticaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Avatica: %w", err)
	}

	// 配置连接池
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Avatica: %w", err)
	}

	return &CalciteClient{
		db:    db,
		cache: cache,
	}, nil
}

// ExecuteQuery 执行查询（带缓存）
func (c *CalciteClient) ExecuteQuery(ctx context.Context, sql string, params ...interface{}) ([]map[string]interface{}, error) {
	// 生成缓存键
	cacheKey := c.generateCacheKey(sql, params...)

	// 检查缓存
	if c.cache != nil {
		if cached, err := c.cache.Get(ctx, cacheKey); err == nil {
			if result, ok := cached.([]map[string]interface{}); ok {
				return result, nil
			}
		}
	}

	// 执行查询
	result, err := c.executeQueryNoCache(ctx, sql, params...)
	if err != nil {
		return nil, err
	}

	// 缓存结果 (TTL 5分钟)
	if c.cache != nil {
		_ = c.cache.Set(ctx, cacheKey, result, 5*time.Minute)
	}

	return result, nil
}

// ExecuteQueryNoCache 执行查询（不使用缓存）
func (c *CalciteClient) ExecuteQueryNoCache(ctx context.Context, sql string, params ...interface{}) ([]map[string]interface{}, error) {
	return c.executeQueryNoCache(ctx, sql, params...)
}

// executeQueryNoCache 内部方法：执行查询
func (c *CalciteClient) executeQueryNoCache(ctx context.Context, sql string, params ...interface{}) ([]map[string]interface{}, error) {
	rows, err := c.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	return c.parseRows(rows)
}

// parseRows 解析SQL结果为map数组
func (c *CalciteClient) parseRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var result []map[string]interface{}

	for rows.Next() {
		// 创建扫描目标
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 构建结果Map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// 处理字节数组
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return result, nil
}

// generateCacheKey 生成缓存键
func (c *CalciteClient) generateCacheKey(sql string, params ...interface{}) string {
	// 使用MD5生成缓存键
	hasher := md5.New()
	hasher.Write([]byte(sql))
	for _, param := range params {
		hasher.Write([]byte(fmt.Sprintf("%v", param)))
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("query:%s", hash)
}

// InvalidateCache 清除指定SQL的缓存
func (c *CalciteClient) InvalidateCache(ctx context.Context, sql string, params ...interface{}) error {
	if c.cache == nil {
		return nil
	}
	cacheKey := c.generateCacheKey(sql, params...)
	return c.cache.Delete(ctx, cacheKey)
}

// Close 关闭连接
func (c *CalciteClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Stats 获取连接池统计信息
func (c *CalciteClient) Stats() sql.DBStats {
	if c.db != nil {
		return c.db.Stats()
	}
	return sql.DBStats{}
}

// Ping 测试连接
func (c *CalciteClient) Ping(ctx context.Context) error {
	if c.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	return c.db.PingContext(ctx)
}
