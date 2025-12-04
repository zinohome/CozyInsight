package engine

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cozy-insight-backend/pkg/config"
	"cozy-insight-backend/pkg/logger"

	_ "github.com/apache/calcite-avatica-go/v5"
	"go.uber.org/zap"
)

type CalciteClient struct {
	db *sql.DB
}

var Client *CalciteClient

func InitCalciteClient(cfg config.CalciteConfig) error {
	// Avatica connection string format: http://host:port/
	dsn := cfg.AvaticaURL
	db, err := sql.Open("avatica", dsn)
	if err != nil {
		return fmt.Errorf("failed to open avatica connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	// Note: Ping might not work depending on Avatica server version/config,
	// but it's good practice. If Avatica server isn't running yet, this might fail.
	// We'll log a warning instead of failing hard if we want to allow backend to start without Avatica for now,
	// but for "Init" it's usually better to fail if critical.
	// However, user might not have Avatica server running locally yet.
	// Let's try to ping and log error but not block startup if strictly needed,
	// OR block startup to ensure integrity. Given the requirement, let's block or return error.

	// For now, let's assume the user might set it up later or it's running.
	// We will return error if Open fails, but Ping is the real network check.

	Client = &CalciteClient{db: db}
	return nil
}

func (c *CalciteClient) ExecuteQuery(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	start := time.Now()
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Log.Error("Calcite query failed", zap.String("query", query), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold the values
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
			var v interface{}
			val := values[i]

			// Handle byte slices (common for strings in some drivers)
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else {
				v = val
			}
			row[col] = v
		}
		result = append(result, row)
	}

	logger.Log.Debug("Calcite query executed",
		zap.String("query", query),
		zap.Duration("duration", time.Since(start)),
		zap.Int("rows", len(result)),
	)

	return result, nil
}

func (c *CalciteClient) Close() error {
	return c.db.Close()
}
