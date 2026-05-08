package engine

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type fileConnector struct {
	db        *sql.DB
	filePath  string
	fileType  string
	tableName string
}

func (c *fileConnector) Connect(configJSON string) error {
	if c.db != nil {
		return nil
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return fmt.Errorf("invalid config json: %w", err)
	}
	filePath, ok := cfg["file_path"].(string)
	if !ok || filePath == "" {
		return fmt.Errorf("missing or invalid file_path")
	}
	fileType, _ := cfg["file_type"].(string)
	if fileType == "" {
		fileType = strings.ToLower(filepath.Ext(filePath))
		if fileType != "" {
			fileType = fileType[1:] // remove leading dot
		}
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return fmt.Errorf("open sqlite memory failed: %w", err)
	}

	tableName := "data"
	switch fileType {
	case "xlsx", "xls", "excel":
		if err := parseExcelToSQLite(filePath, db, tableName); err != nil {
			db.Close()
			return fmt.Errorf("parse excel failed: %w", err)
		}
	case "csv":
		if err := parseCSVToSQLite(filePath, db, tableName); err != nil {
			db.Close()
			return fmt.Errorf("parse csv failed: %w", err)
		}
	default:
		db.Close()
		return fmt.Errorf("unsupported file type: %s", fileType)
	}

	c.db = db
	c.filePath = filePath
	c.fileType = fileType
	c.tableName = tableName
	return nil
}

func (c *fileConnector) Close() error {
	if c.db == nil {
		return nil
	}
	err := c.db.Close()
	c.db = nil
	return err
}

func (c *fileConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
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

func (c *fileConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]ColumnInfo, error) {
	if c.db == nil {
		return nil, fmt.Errorf("not connected")
	}
	query := fmt.Sprintf("PRAGMA table_info(%s)", QuoteIdentifier(c.tableName, "sqlite"))
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

func parseExcelToSQLite(filePath string, db *sql.DB, tableName string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("open excel failed: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("get rows failed: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("empty excel file")
	}

	headers := rows[0]
	colDefs := make([]string, len(headers))
	for i, h := range headers {
		colDefs[i] = fmt.Sprintf("%s TEXT", QuoteIdentifier(sanitizeName(h, i), "sqlite"))
	}

	createSQL := fmt.Sprintf("CREATE TABLE %s (%s)", QuoteIdentifier(tableName, "sqlite"), strings.Join(colDefs, ", "))
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	placeholders := make([]string, len(headers))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s VALUES (%s)", QuoteIdentifier(tableName, "sqlite"), strings.Join(placeholders, ", "))
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("prepare insert failed: %w", err)
	}
	defer stmt.Close()

	for i, row := range rows[1:] {
		values := make([]interface{}, len(headers))
		for j := range headers {
			if j < len(row) {
				values[j] = row[j]
			} else {
				values[j] = ""
			}
		}
		if _, err := stmt.Exec(values...); err != nil {
			return fmt.Errorf("insert row %d failed: %w", i+1, err)
		}
	}
	return nil
}

func parseCSVToSQLite(filePath string, db *sql.DB, tableName string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open csv failed: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("read csv failed: %w", err)
	}
	if len(records) == 0 {
		return fmt.Errorf("empty csv file")
	}

	headers := records[0]
	colDefs := make([]string, len(headers))
	for i, h := range headers {
		colDefs[i] = fmt.Sprintf("%s TEXT", QuoteIdentifier(sanitizeName(h, i), "sqlite"))
	}

	createSQL := fmt.Sprintf("CREATE TABLE %s (%s)", QuoteIdentifier(tableName, "sqlite"), strings.Join(colDefs, ", "))
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	placeholders := make([]string, len(headers))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s VALUES (%s)", QuoteIdentifier(tableName, "sqlite"), strings.Join(placeholders, ", "))
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("prepare insert failed: %w", err)
	}
	defer stmt.Close()

	for i, row := range records[1:] {
		values := make([]interface{}, len(headers))
		for j := range headers {
			if j < len(row) {
				values[j] = row[j]
			} else {
				values[j] = ""
			}
		}
		if _, err := stmt.Exec(values...); err != nil {
			return fmt.Errorf("insert row %d failed: %w", i+1, err)
		}
	}
	return nil
}

func sanitizeName(name string, idx int) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	if name == "" {
		name = "col" + strconv.Itoa(idx)
	}
	return name
}
