package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// DatasourceConnector 数据源连接器，用于测试各类数据源连接
type DatasourceConnector struct {
	calcite *CalciteClient
}

// NewDatasourceConnector 创建数据源连接器
func NewDatasourceConnector(calcite *CalciteClient) *DatasourceConnector {
	return &DatasourceConnector{
		calcite: calcite,
	}
}

// ConnectionConfig ... 数据源连接配置
type ConnectionConfig struct {
	Type     string                 `json:"type"`
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	Database string                 `json:"database"`
	Username string                 `json:"username"`
	Password string                 `json:"password"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

// ConnectionTestResult 连接测试结果
type ConnectionTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Latency int64  `json:"latency"` // 毫秒
}

// TestConnection 测试数据源连接
func (dc *DatasourceConnector) TestConnection(ctx context.Context, configJSON string) (*ConnectionTestResult, error) {
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("invalid connection config: %w", err)
	}

	startTime := time.Now()

	var err error
	switch config.Type {
	case "mysql":
		err = dc.testMySQL(ctx, &config)
	case "postgresql", "postgres":
		err = dc.testPostgreSQL(ctx, &config)
	case "clickhouse":
		err = dc.testClickHouse(ctx, &config)
	case "oracle":
		err = dc.testOracle(ctx, &config)
	case "sqlserver", "mssql":
		err = dc.testSQLServer(ctx, &config)
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", config.Type)
	}

	latency := time.Since(startTime).Milliseconds()

	result := &ConnectionTestResult{
		Success: err == nil,
		Latency: latency,
	}

	if err != nil {
		result.Message = fmt.Sprintf("连接失败: %s", err.Error())
	} else {
		result.Message = fmt.Sprintf("连接成功！延迟: %dms", latency)
	}

	return result, nil
}

// testMySQL 测试 MySQL 连接
func (dc *DatasourceConnector) testMySQL(ctx context.Context, config *ConnectionConfig) error {
	// 构建测试 SQL
	// 注意：实际的连接是通过 Avatica Server 建立的
	// 这里我们假设 Avatica Server 已配置好对应的数据源
	// 或者我们需要动态注册数据源到 Avatica

	// 简化版本：直接执行一个简单查询
	testSQL := "SELECT 1 as test"

	// 如果需要指定数据库，可以在 SQL 中指定
	if config.Database != "" {
		testSQL = fmt.Sprintf("SELECT 1 as test FROM %s.dual", config.Database)
	}

	// 通过 Calcite 执行查询
	_, err := dc.calcite.ExecuteQueryNoCache(ctx, testSQL)
	return err
}

// testPostgreSQL 测试 PostgreSQL 连接
func (dc *DatasourceConnector) testPostgreSQL(ctx context.Context, config *ConnectionConfig) error {
	testSQL := "SELECT 1 as test"
	_, err := dc.calcite.ExecuteQueryNoCache(ctx, testSQL)
	return err
}

// testClickHouse 测试 ClickHouse 连接
func (dc *DatasourceConnector) testClickHouse(ctx context.Context, config *ConnectionConfig) error {
	testSQL := "SELECT 1 as test"
	_, err := dc.calcite.ExecuteQueryNoCache(ctx, testSQL)
	return err
}

// testOracle 测试 Oracle 连接
func (dc *DatasourceConnector) testOracle(ctx context.Context, config *ConnectionConfig) error {
	testSQL := "SELECT 1 FROM dual"
	_, err := dc.calcite.ExecuteQueryNoCache(ctx, testSQL)
	return err
}

// testSQLServer 测试 SQL Server 连接
func (dc *DatasourceConnector) testSQLServer(ctx context.Context, config *ConnectionConfig) error {
	testSQL := "SELECT 1 as test"
	_, err := dc.calcite.ExecuteQueryNoCache(ctx, testSQL)
	return err
}

// GetDatabaseList 获取数据库列表
func (dc *DatasourceConnector) GetDatabaseList(ctx context.Context, configJSON string) ([]string, error) {
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("invalid connection config: %w", err)
	}

	var query string
	switch config.Type {
	case "mysql":
		query = "SHOW DATABASES"
	case "postgresql", "postgres":
		query = "SELECT datname FROM pg_database WHERE datistemplate = false"
	case "sqlserver", "mssql":
		query = "SELECT name FROM sys.databases"
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", config.Type)
	}

	rows, err := dc.calcite.ExecuteQueryNoCache(ctx, query)
	if err != nil {
		return nil, err
	}

	var databases []string
	for _, row := range rows {
		for _, val := range row {
			if str, ok := val.(string); ok {
				databases = append(databases, str)
			}
		}
	}

	return databases, nil
}

// GetTableList 获取表列表
func (dc *DatasourceConnector) GetTableList(ctx context.Context, configJSON string, database string) ([]string, error) {
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("invalid connection config: %w", err)
	}

	var query string
	switch config.Type {
	case "mysql":
		if database == "" {
			database = config.Database
		}
		query = fmt.Sprintf("SHOW TABLES FROM %s", database)
	case "postgresql", "postgres":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
	case "sqlserver", "mssql":
		query = "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE'"
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", config.Type)
	}

	rows, err := dc.calcite.ExecuteQueryNoCache(ctx, query)
	if err != nil {
		return nil, err
	}

	var tables []string
	for _, row := range rows {
		for _, val := range row {
			if str, ok := val.(string); ok {
				tables = append(tables, str)
			}
		}
	}

	return tables, nil
}

// GetTableSchema 获取表结构
func (dc *DatasourceConnector) GetTableSchema(ctx context.Context, configJSON string, database, table string) ([]map[string]interface{}, error) {
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("invalid connection config: %w", err)
	}

	var query string
	switch config.Type {
	case "mysql":
		if database == "" {
			database = config.Database
		}
		query = fmt.Sprintf("DESCRIBE %s.%s", database, table)
	case "postgresql", "postgres":
		query = fmt.Sprintf(`
			SELECT column_name, data_type, character_maximum_length, is_nullable
			FROM information_schema.columns
			WHERE table_name = '%s'
			ORDER BY ordinal_position
		`, table)
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", config.Type)
	}

	return dc.calcite.ExecuteQueryNoCache(ctx, query)
}

// DirectQuery 直接执行 SQL 查询（用于自定义 SQL 数据集）
func (dc *DatasourceConnector) DirectQuery(ctx context.Context, configJSON string, query string, useCache bool) ([]map[string]interface{}, error) {
	if useCache {
		return dc.calcite.ExecuteQuery(ctx, query)
	}
	return dc.calcite.ExecuteQueryNoCache(ctx, query)
}

// ParseConnectionString 从数据库连接字符串创建 ConnectionConfig
func ParseConnectionString(connStr string) (*ConnectionConfig, error) {
	// 这里简化处理，实际应该解析标准的连接字符串格式
	// 例如: mysql://user:pass@host:port/database
	return nil, fmt.Errorf("not implemented")
}

// BuildJDBCUrl 构建 JDBC 连接 URL
func (config *ConnectionConfig) BuildJDBCUrl() string {
	switch config.Type {
	case "mysql":
		return fmt.Sprintf("jdbc:mysql://%s:%d/%s?useSSL=false&serverTimezone=UTC",
			config.Host, config.Port, config.Database)
	case "postgresql", "postgres":
		return fmt.Sprintf("jdbc:postgresql://%s:%d/%s",
			config.Host, config.Port, config.Database)
	case "clickhouse":
		return fmt.Sprintf("jdbc:clickhouse://%s:%d/%s",
			config.Host, config.Port, config.Database)
	case "oracle":
		return fmt.Sprintf("jdbc:oracle:thin:@%s:%d:%s",
			config.Host, config.Port, config.Database)
	case "sqlserver", "mssql":
		return fmt.Sprintf("jdbc:sqlserver://%s:%d;databaseName=%s",
			config.Host, config.Port, config.Database)
	default:
		return ""
	}
}

// Validate 验证连接配置
func (config *ConnectionConfig) Validate() error {
	if config.Type == "" {
		return fmt.Errorf("datasource type is required")
	}
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}
	if config.Port == 0 {
		return fmt.Errorf("port is required")
	}
	return nil
}
