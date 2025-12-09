package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DatasourceConnector 数据源连接器,用于测试连接和获取元数据
type DatasourceConnector struct{}

// ConnectionTestResult 连接测试结果
type ConnectionTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Time    int64  `json:"time"` // 响应时间(毫秒)
}

// DatasourceConfig 数据源配置
type DatasourceConfig struct {
	Type     string `json:"type"`     // mysql, postgresql, clickhouse, etc.
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Charset  string `json:"charset"`
}

func NewDatasourceConnector() *DatasourceConnector {
	return &DatasourceConnector{}
}

// TestConnection 测试数据源连接
func (c *DatasourceConnector) TestConnection(ctx context.Context, configJSON string) (*ConnectionTestResult, error) {
	start := time.Now()
	result := &ConnectionTestResult{
		Success: false,
	}

	// 解析配置
	var config DatasourceConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		result.Message = fmt.Sprintf("Invalid configuration: %v", err)
		result.Time = time.Since(start).Milliseconds()
		return result, nil
	}

	// 构建连接字符串
	dsn, err := c.buildDSN(&config)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to build DSN: %v", err)
		result.Time = time.Since(start).Milliseconds()
		return result, nil
	}

	// 尝试连接
	db, err := sql.Open(c.getDriverName(config.Type), dsn)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to open connection: %v", err)
		result.Time = time.Since(start).Milliseconds()
		return result, nil
	}
	defer db.Close()

	// 设置连接超时
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Ping测试
	if err := db.PingContext(ctx); err != nil {
		result.Message = fmt.Sprintf("Connection failed: %v", err)
		result.Time = time.Since(start).Milliseconds()
		return result, nil
	}

	result.Success = true
	result.Message = "Connection successful"
	result.Time = time.Since(start).Milliseconds()
	return result, nil
}

// GetDatabaseList 获取数据库列表
func (c *DatasourceConnector) GetDatabaseList(ctx context.Context, configJSON string) ([]string, error) {
	var config DatasourceConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	dsn, err := c.buildDSN(&config)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(c.getDriverName(config.Type), dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var query string
	switch config.Type {
	case "mysql":
		query = "SHOW DATABASES"
	case "postgresql":
		query = "SELECT datname FROM pg_database WHERE datistemplate = false"
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", config.Type)
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		databases = append(databases, name)
	}

	return databases, rows.Err()
}

// GetTableList 获取表列表
func (c *DatasourceConnector) GetTableList(ctx context.Context, configJSON, database string) ([]string, error) {
	var config DatasourceConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// 覆盖数据库
	config.Database = database

	dsn, err := c.buildDSN(&config)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(c.getDriverName(config.Type), dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var query string
	switch config.Type {
	case "mysql":
		query = "SHOW TABLES"
	case "postgresql":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", config.Type)
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	return tables, rows.Err()
}

// GetTableSchema 获取表结构
func (c *DatasourceConnector) GetTableSchema(ctx context.Context, configJSON, database, table string) ([]map[string]interface{}, error) {
	var config DatasourceConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	config.Database = database

	dsn, err := c.buildDSN(&config)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(c.getDriverName(config.Type), dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var query string
	var args []interface{}

	switch config.Type {
	case "mysql":
		query = "DESCRIBE " + table
	case "postgresql":
		query = `SELECT column_name, data_type, is_nullable 
                 FROM information_schema.columns 
                 WHERE table_name = $1 
                 ORDER BY ordinal_position`
		args = []interface{}{table}
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", config.Type)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var schema []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		schema = append(schema, row)
	}

	return schema, rows.Err()
}

// buildDSN 构建数据源连接字符串
func (c *DatasourceConnector) buildDSN(config *DatasourceConfig) (string, error) {
	switch config.Type {
	case "mysql":
		charset := config.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
			charset), nil

	case "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host,
			config.Port,
			config.Username,
			config.Password,
			config.Database), nil

	default:
		return "", fmt.Errorf("unsupported datasource type: %s", config.Type)
	}
}

// getDriverName 获取驱动名称
func (c *DatasourceConnector) getDriverName(dsType string) string {
	switch dsType {
	case "mysql":
		return "mysql"
	case "postgresql":
		return "postgres"
	default:
		return dsType
	}
}
