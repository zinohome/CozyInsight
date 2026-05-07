package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// DatasourceConnector abstracts connections to external databases.
type DatasourceConnector interface {
	Connect(configJSON string) error
	Close() error
	Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error)
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
	host, _ := cfg["host"].(string)
	portF, _ := cfg["port"].(float64)
	username, _ := cfg["username"].(string)
	password, _ := cfg["password"].(string)
	database, _ := cfg["database"].(string)
	port := int(portF)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		username, password, host, port, database), nil
}

func (c *mysqlConnector) Connect(configJSON string) error {
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
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *mysqlConnector) Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	rows, err := c.db.QueryContext(ctx, sql, args...)
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
	return cols, fmt.Errorf("row iteration failed: %w", rows.Err())
}

type postgresqlConnector struct {
	db *sql.DB
}

func (c *postgresqlConnector) buildDSN(configJSON string) (string, error) {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", fmt.Errorf("invalid config json: %w", err)
	}
	host, _ := cfg["host"].(string)
	portF, _ := cfg["port"].(float64)
	username, _ := cfg["username"].(string)
	password, _ := cfg["password"].(string)
	database, _ := cfg["database"].(string)
	port := int(portF)
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, database), nil
}

func (c *postgresqlConnector) Connect(configJSON string) error {
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
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *postgresqlConnector) Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	rows, err := c.db.QueryContext(ctx, sql, args...)
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
	return cols, fmt.Errorf("row iteration failed: %w", rows.Err())
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
	return result, fmt.Errorf("row iteration failed: %w", rows.Err())
}
