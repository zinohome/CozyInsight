package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
)

// DatasourceConnector abstracts connections to external databases.
type DatasourceConnector interface {
	Connect(configJSON string) error
	Close() error
	Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
	GetColumns(ctx context.Context, dbName, tableName string) ([]ColumnInfo, error)
}

// ColumnInfo holds metadata for a single column.
type ColumnInfo struct {
	Name      string
	Type      string
	Length    int
	Precision int
	Scale     int
}

// NewConnector returns a connector for the given datasource type.
func NewConnector(dsType string) (DatasourceConnector, error) {
	switch dsType {
	case "mysql":
		return &mysqlConnector{}, nil
	case "postgresql":
		return &postgresqlConnector{}, nil
	case "sqlite":
		return &sqliteConnector{}, nil
	case "clickhouse":
		return &clickhouseConnector{}, nil
	case "sqlserver":
		return &sqlserverConnector{}, nil
	case "excel", "csv":
		return &fileConnector{}, nil
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", dsType)
	}
}

type mysqlConnector struct {
	db *sql.DB
}

func (c *mysqlConnector) buildDSN(configJSON string) (string, error) {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", fmt.Errorf("invalid config json: %w", err)
	}
	host, ok := cfg["host"].(string)
	if !ok || host == "" {
		return "", fmt.Errorf("missing or invalid required field: host")
	}
	portF, ok := cfg["port"].(float64)
	if !ok || portF <= 0 {
		return "", fmt.Errorf("missing or invalid required field: port")
	}
	username, ok := cfg["username"].(string)
	if !ok || username == "" {
		return "", fmt.Errorf("missing or invalid required field: username")
	}
	password, _ := cfg["password"].(string)
	database, ok := cfg["database"].(string)
	if !ok || database == "" {
		return "", fmt.Errorf("missing or invalid required field: database")
	}
	port := int(portF)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		username, password, host, port, database), nil
}

func (c *mysqlConnector) Connect(configJSON string) error {
	if c.db != nil {
		return nil // already connected
	}
	dsn, err := c.buildDSN(configJSON)
	if err != nil {
		return err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping mysql failed: %w", err)
	}
	c.db = db
	return nil
}

func (c *mysqlConnector) Close() error {
	if c.db == nil {
		return nil
	}
	err := c.db.Close()
	c.db = nil
	return err
}

func (c *mysqlConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (c *mysqlConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]ColumnInfo, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH, NUMERIC_PRECISION, NUMERIC_SCALE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`
	rows, err := c.db.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}
	defer rows.Close()

	var cols []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var maxLen, prec, scale sql.NullInt64
		if err := rows.Scan(&col.Name, &col.Type, &maxLen, &prec, &scale); err != nil {
			return nil, fmt.Errorf("scan column info failed: %w", err)
		}
		if maxLen.Valid {
			col.Length = int(maxLen.Int64)
		}
		if prec.Valid {
			col.Precision = int(prec.Int64)
		}
		if scale.Valid {
			col.Scale = int(scale.Int64)
		}
		cols = append(cols, col)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}
	return cols, nil
}

type postgresqlConnector struct {
	db *sql.DB
}

func (c *postgresqlConnector) buildDSN(configJSON string) (string, error) {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", fmt.Errorf("invalid config json: %w", err)
	}
	host, ok := cfg["host"].(string)
	if !ok || host == "" {
		return "", fmt.Errorf("missing or invalid required field: host")
	}
	portF, ok := cfg["port"].(float64)
	if !ok || portF <= 0 {
		return "", fmt.Errorf("missing or invalid required field: port")
	}
	username, ok := cfg["username"].(string)
	if !ok || username == "" {
		return "", fmt.Errorf("missing or invalid required field: username")
	}
	password, _ := cfg["password"].(string)
	database, ok := cfg["database"].(string)
	if !ok || database == "" {
		return "", fmt.Errorf("missing or invalid required field: database")
	}
	sslmode, _ := cfg["sslmode"].(string)
	if sslmode == "" {
		sslmode = "prefer"
	}
	port := int(portF)
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, username, password, database, sslmode), nil
}

func (c *postgresqlConnector) Connect(configJSON string) error {
	if c.db != nil {
		return nil // already connected
	}
	dsn, err := c.buildDSN(configJSON)
	if err != nil {
		return err
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open postgresql failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping postgresql failed: %w", err)
	}
	c.db = db
	return nil
}

func (c *postgresqlConnector) Close() error {
	if c.db == nil {
		return nil
	}
	err := c.db.Close()
	c.db = nil
	return err
}

func (c *postgresqlConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (c *postgresqlConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]ColumnInfo, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	query := `
		SELECT column_name, data_type, character_maximum_length, numeric_precision, numeric_scale
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position`
	rows, err := c.db.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}
	defer rows.Close()

	var cols []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var maxLen, prec, scale sql.NullInt64
		if err := rows.Scan(&col.Name, &col.Type, &maxLen, &prec, &scale); err != nil {
			return nil, fmt.Errorf("scan column info failed: %w", err)
		}
		if maxLen.Valid {
			col.Length = int(maxLen.Int64)
		}
		if prec.Valid {
			col.Precision = int(prec.Int64)
		}
		if scale.Valid {
			col.Scale = int(scale.Int64)
		}
		cols = append(cols, col)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}
	return cols, nil
}

type sqliteConnector struct {
	db *sql.DB
}

func (c *sqliteConnector) Connect(configJSON string) error {
	if c.db != nil {
		return nil
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return fmt.Errorf("invalid config json: %w", err)
	}
	dbName, ok := cfg["database"].(string)
	if !ok || dbName == "" {
		return fmt.Errorf("missing or invalid required field: database")
	}
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return fmt.Errorf("open sqlite3 failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping sqlite3 failed: %w", err)
	}
	c.db = db
	return nil
}

func (c *sqliteConnector) Close() error {
	if c.db == nil {
		return nil
	}
	err := c.db.Close()
	c.db = nil
	return err
}

func (c *sqliteConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (c *sqliteConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]ColumnInfo, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	query := fmt.Sprintf("PRAGMA table_info(%s)", QuoteIdentifier(tableName, "sqlite"))
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}
	defer rows.Close()

	var cols []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var cid int
		var notNull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &col.Name, &col.Type, &notNull, &dflt, &pk); err != nil {
			return nil, fmt.Errorf("scan column info failed: %w", err)
		}
		cols = append(cols, col)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}
	return cols, nil
}

type clickhouseConnector struct {
	db *sql.DB
}

func (c *clickhouseConnector) Connect(configJSON string) error {
	if c.db != nil {
		return nil
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return fmt.Errorf("invalid config json: %w", err)
	}
	host, ok := cfg["host"].(string)
	if !ok || host == "" {
		return fmt.Errorf("missing or invalid required field: host")
	}
	portF, ok := cfg["port"].(float64)
	if !ok || portF <= 0 {
		return fmt.Errorf("missing or invalid required field: port")
	}
	username, ok := cfg["username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("missing or invalid required field: username")
	}
	password, _ := cfg["password"].(string)
	database, ok := cfg["database"].(string)
	if !ok || database == "" {
		return fmt.Errorf("missing or invalid required field: database")
	}
	port := int(portF)
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s", username, password, host, port, database)
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return fmt.Errorf("open clickhouse failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping clickhouse failed: %w", err)
	}
	c.db = db
	return nil
}

func (c *clickhouseConnector) Close() error {
	if c.db == nil {
		return nil
	}
	err := c.db.Close()
	c.db = nil
	return err
}

func (c *clickhouseConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (c *clickhouseConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]ColumnInfo, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	query := `
		SELECT name, type, 0, 0, 0
		FROM system.columns
		WHERE database = ? AND table = ?
		ORDER BY position`
	rows, err := c.db.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}
	defer rows.Close()

	var cols []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var length, prec, scale int
		if err := rows.Scan(&col.Name, &col.Type, &length, &prec, &scale); err != nil {
			return nil, fmt.Errorf("scan column info failed: %w", err)
		}
		col.Length = length
		col.Precision = prec
		col.Scale = scale
		cols = append(cols, col)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}
	return cols, nil
}

type sqlserverConnector struct {
	db *sql.DB
}

func (c *sqlserverConnector) buildDSN(configJSON string) (string, error) {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", fmt.Errorf("invalid config json: %w", err)
	}
	host, ok := cfg["host"].(string)
	if !ok || host == "" {
		return "", fmt.Errorf("missing or invalid required field: host")
	}
	portF, ok := cfg["port"].(float64)
	if !ok || portF <= 0 {
		return "", fmt.Errorf("missing or invalid required field: port")
	}
	username, ok := cfg["username"].(string)
	if !ok || username == "" {
		return "", fmt.Errorf("missing or invalid required field: username")
	}
	password, _ := cfg["password"].(string)
	database, ok := cfg["database"].(string)
	if !ok || database == "" {
		return "", fmt.Errorf("missing or invalid required field: database")
	}
	port := int(portF)
	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		username, password, host, port, database), nil
}

func (c *sqlserverConnector) Connect(configJSON string) error {
	if c.db != nil {
		return nil
	}
	dsn, err := c.buildDSN(configJSON)
	if err != nil {
		return err
	}
	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return fmt.Errorf("open sqlserver failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping sqlserver failed: %w", err)
	}
	c.db = db
	return nil
}

func (c *sqlserverConnector) Close() error {
	if c.db == nil {
		return nil
	}
	err := c.db.Close()
	c.db = nil
	return err
}

func (c *sqlserverConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (c *sqlserverConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]ColumnInfo, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH, NUMERIC_PRECISION, NUMERIC_SCALE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_CATALOG = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`
	rows, err := c.db.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}
	defer rows.Close()

	var cols []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var maxLen, prec, scale sql.NullInt64
		if err := rows.Scan(&col.Name, &col.Type, &maxLen, &prec, &scale); err != nil {
			return nil, fmt.Errorf("scan column info failed: %w", err)
		}
		if maxLen.Valid {
			col.Length = int(maxLen.Int64)
		}
		if prec.Valid {
			col.Precision = int(prec.Int64)
		}
		if scale.Valid {
			col.Scale = int(scale.Int64)
		}
		cols = append(cols, col)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}
	return cols, nil
}

func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("get column types failed: %w", err)
	}

	// Guard: if lengths mismatch, skip type-based conversion
	typesAvailable := len(columnTypes) == len(columns)

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scan row failed: %w", err)
		}

		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				if !typesAvailable || i >= len(columnTypes) {
					row[col] = string(b)
					continue
				}
				switch columnTypes[i].DatabaseTypeName() {
				case "INT", "BIGINT", "SMALLINT", "TINYINT", "INTEGER", "INT4", "INT8":
					if n, err := strconv.ParseInt(string(b), 10, 64); err == nil {
						row[col] = n
					} else {
						row[col] = string(b)
					}
				case "FLOAT", "DOUBLE", "DECIMAL", "NUMERIC", "REAL", "FLOAT4", "FLOAT8":
					if n, err := strconv.ParseFloat(string(b), 64); err == nil {
						row[col] = n
					} else {
						row[col] = string(b)
					}
				default:
					row[col] = string(b)
				}
			} else {
				row[col] = val
			}
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}
	return result, nil
}
