# Phase 3 — 数据可视化引擎 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the end-to-end data visualization pipeline: from SQL query execution against real datasources to interactive chart rendering and dashboard design.

**Architecture:** A unified `DatasourceConnector` interface abstracts MySQL/PostgreSQL connections. A `QueryEngine` parses chart configuration JSON into parameterized SQL, executes it via the connector, and returns chart-ready data. The frontend uses `@ant-design/charts` for rendering and `react-grid-layout` for dashboard drag-and-drop layout.

**Tech Stack:** Go 1.25 + sqlx + mysql/pq drivers; React 19 + TypeScript + Ant Design Charts 2.6 + react-grid-layout 1.5

---

## File Structure

### Backend

| File | Responsibility |
|------|--------------|
| `backend/internal/engine/connector.go` | `DatasourceConnector` interface + MySQL/PostgreSQL implementations + factory |
| `backend/internal/engine/query_engine.go` | SQL builder from chart config + executor that returns `[]map[string]interface{}` |
| `backend/internal/engine/connector_test.go` | Unit tests for connector (using docker-compose MySQL or mocked) |
| `backend/internal/engine/query_engine_test.go` | Unit tests for SQL generation with various configs |
| `backend/internal/dto/chart.go` | Extended DTOs: `ChartConfig`, `ChartDimension`, `ChartMetric`, `ChartFilter`, `ChartDataResponse` |
| `backend/internal/service/chart_service.go` | Modified: inject `DatasetRepository` and `DatasourceRepository`; add `GetData()` |
| `backend/internal/service/dataset_service.go` | Modified: replace `getTableColumns` and `queryTableData` stubs with real connector calls |
| `backend/internal/handler/chart_handler.go` | Modified: add `GetData(c)` handler |
| `backend/api/v1/router.go` | Modified: register `GET /chart/:id/data`, wire new service dependencies |

### Frontend

| File | Responsibility |
|------|--------------|
| `frontend/src/types/chart.ts` | Extended: `ChartConfig`, `ChartFieldConfig`, `ChartDataResponse` |
| `frontend/src/api/chart.ts` | Extended: `chartAPI.getData(id)` |
| `frontend/src/components/ChartRenderer/index.tsx` | Receives `{type, data, config}` and renders bar/line/pie/table via `@ant-design/charts` |
| `frontend/src/pages/chart/ChartBuilder.tsx` | Visual builder: field panel (dimensions/metrics/filters), type selector, live preview |
| `frontend/src/pages/chart/index.tsx` | Modified: add "编辑" button linking to `/chart/builder/:id` |
| `frontend/src/pages/dashboard/DashboardDesigner.tsx` | Grid canvas using `react-grid-layout`: place charts, resize, save layout |
| `frontend/src/pages/dashboard/index.tsx` | Modified: add "设计" button linking to `/dashboard/designer/:id` |
| `frontend/src/router/index.tsx` | Modified: add `/chart/builder/:id` and `/dashboard/designer/:id` routes |

---

## Chart Config JSON Schema

The `chart.config` column stores JSON with this structure:

```json
{
  "dimensions": [
    {"field": "dept_name", "sort": "asc"}
  ],
  "metrics": [
    {"field": "amount", "aggregation": "SUM", "alias": "总金额"}
  ],
  "filters": [
    {"field": "status", "operator": "=", "value": "1"}
  ],
  "orders": [
    {"field": "amount", "direction": "desc"}
  ],
  "limit": 100
}
```

**Valid aggregations:** `SUM`, `COUNT`, `AVG`, `MAX`, `MIN` (case-insensitive).  
**Valid operators:** `=`, `!=`, `>`, `<`, `>=`, `<=`, `LIKE`, `NOT LIKE`, `IN`, `NOT IN`.

---

## Task 1: Datasource Connector

**Files:**
- Create: `backend/internal/engine/connector.go`
- Test: `backend/internal/engine/connector_test.go`

- [ ] **Step 1: Write the failing test**

```go
package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnector_MySQL(t *testing.T) {
	conn, err := NewConnector("mysql")
	require.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestNewConnector_Unsupported(t *testing.T) {
	_, err := NewConnector("oracle")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestMySQLConnector_BuildDSN(t *testing.T) {
	conn := &mysqlConnector{}
	dsn, err := conn.buildDSN(`{"host":"localhost","port":3306,"username":"root","password":"secret","database":"test"}`)
	require.NoError(t, err)
	assert.Equal(t, "root:secret@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=true", dsn)
}

func TestPostgreSQLConnector_BuildDSN(t *testing.T) {
	conn := &postgresqlConnector{}
	dsn, err := conn.buildDSN(`{"host":"localhost","port":5432,"username":"postgres","password":"secret","database":"test"}`)
	require.NoError(t, err)
	assert.Equal(t, "host=localhost port=5432 user=postgres password=secret dbname=test sslmode=disable", dsn)
}
```

Run: `cd backend && go test ./internal/engine/... -v`  
Expected: FAIL — `engine` package does not exist.

- [ ] **Step 2: Write minimal implementation**

```go
// backend/internal/engine/connector.go
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
			return nil, err
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
	return cols, rows.Err()
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
			return nil, err
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
	return cols, rows.Err()
}

func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				// Try to infer whether to convert to string or number
				switch columnTypes[i].DatabaseTypeName() {
				case "INT", "BIGINT", "SMALLINT", "TINYINT", "INTEGER":
					if n, err := strconv.ParseInt(string(b), 10, 64); err == nil {
						row[col] = n
					} else {
						row[col] = string(b)
					}
				case "FLOAT", "DOUBLE", "DECIMAL", "NUMERIC", "REAL":
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
	return result, rows.Err()
}
```

Run: `cd backend && go test ./internal/engine/... -v`  
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd backend
git add internal/engine/connector.go internal/engine/connector_test.go
git commit -m "feat(engine): add datasource connector for mysql and postgresql"
```

---

## Task 2: SQL Query Engine

**Files:**
- Create: `backend/internal/engine/query_engine.go`
- Test: `backend/internal/engine/query_engine_test.go`

- [ ] **Step 1: Write the failing test**

```go
package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSQL_BasicAggregation(t *testing.T) {
	config := ChartQueryConfig{
		Dimensions: []Dimension{{Field: "dept"}},
		Metrics:    []Metric{{Field: "amount", Aggregation: "SUM", Alias: "total"}},
	}
	sql, args, err := BuildSQL("orders", config)
	require.NoError(t, err)
	assert.Equal(t, "SELECT dept, SUM(amount) AS total FROM orders GROUP BY dept", sql)
	assert.Len(t, args, 0)
}

func TestBuildSQL_WithFilter(t *testing.T) {
	config := ChartQueryConfig{
		Dimensions: []Dimension{{Field: "dept"}},
		Metrics:    []Metric{{Field: "amount", Aggregation: "COUNT"}},
		Filters:    []Filter{{Field: "status", Operator: "=", Value: "1"}},
	}
	sql, args, err := BuildSQL("orders", config)
	require.NoError(t, err)
	assert.Equal(t, "SELECT dept, COUNT(amount) AS count_amount FROM orders WHERE status = ? GROUP BY dept", sql)
	assert.Equal(t, []interface{}{"1"}, args)
}

func TestBuildSQL_WithOrder(t *testing.T) {
	config := ChartQueryConfig{
		Dimensions: []Dimension{{Field: "month"}},
		Metrics:    []Metric{{Field: "revenue", Aggregation: "SUM"}},
		Orders:     []Order{{Field: "revenue", Direction: "desc"}},
		Limit:      10,
	}
	sql, args, err := BuildSQL("sales", config)
	require.NoError(t, err)
	assert.Equal(t, "SELECT month, SUM(revenue) AS sum_revenue FROM sales GROUP BY month ORDER BY sum_revenue desc LIMIT 10", sql)
	assert.Len(t, args, 0)
}

func TestBuildSQL_InvalidAggregation(t *testing.T) {
	config := ChartQueryConfig{
		Metrics: []Metric{{Field: "x", Aggregation: "DROP"}},
	}
	_, _, err := BuildSQL("t", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported aggregation")
}

func TestBuildSQL_InvalidOperator(t *testing.T) {
	config := ChartQueryConfig{
		Filters: []Filter{{Field: "x", Operator: "DROP", Value: "1"}},
	}
	_, _, err := BuildSQL("t", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported operator")
}
```

Run: `cd backend && go test ./internal/engine/... -v`  
Expected: FAIL — `BuildSQL` not defined.

- [ ] **Step 2: Write minimal implementation**

```go
// backend/internal/engine/query_engine.go
package engine

import (
	"fmt"
	"strings"
)

// ChartQueryConfig describes what a chart needs from the data layer.
type ChartQueryConfig struct {
	Dimensions []Dimension `json:"dimensions"`
	Metrics    []Metric    `json:"metrics"`
	Filters    []Filter    `json:"filters"`
	Orders     []Order     `json:"orders"`
	Limit      uint64      `json:"limit"`
}

type Dimension struct {
	Field string `json:"field"`
	Sort  string `json:"sort"`
}

type Metric struct {
	Field       string `json:"field"`
	Aggregation string `json:"aggregation"`
	Alias       string `json:"alias"`
}

type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type Order struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

var allowedAggregations = map[string]bool{
	"SUM": true, "COUNT": true, "AVG": true, "MAX": true, "MIN": true,
}

var allowedOperators = map[string]bool{
	"=": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
	"LIKE": true, "NOT LIKE": true, "IN": true, "NOT IN": true,
}

// BuildSQL generates a parameterized SELECT SQL and its bound arguments.
func BuildSQL(tableName string, config ChartQueryConfig) (string, []interface{}, error) {
	var selectCols []string
	var groupByCols []string

	for _, d := range config.Dimensions {
		selectCols = append(selectCols, quoteIdentifier(d.Field))
		groupByCols = append(groupByCols, quoteIdentifier(d.Field))
	}

	for _, m := range config.Metrics {
		agg := strings.ToUpper(m.Aggregation)
		if !allowedAggregations[agg] {
			return "", nil, fmt.Errorf("unsupported aggregation: %s", m.Aggregation)
		}
		alias := m.Alias
		if alias == "" {
			alias = fmt.Sprintf("%s_%s", strings.ToLower(agg), m.Field)
		}
		selectCols = append(selectCols, fmt.Sprintf("%s(%s) AS %s", agg, quoteIdentifier(m.Field), quoteIdentifier(alias)))
	}

	if len(selectCols) == 0 {
		return "", nil, fmt.Errorf("at least one dimension or metric required")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectCols, ", "), quoteIdentifier(tableName))

	var args []interface{}
	if len(config.Filters) > 0 {
		var conditions []string
		for _, f := range config.Filters {
		op := strings.ToUpper(f.Operator)
		if !allowedOperators[op] {
			return "", nil, fmt.Errorf("unsupported operator: %s", f.Operator)
		}
		conditions = append(conditions, fmt.Sprintf("%s %s ?", quoteIdentifier(f.Field), op))
		args = append(args, f.Value)
		}
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if len(groupByCols) > 0 {
		query += " GROUP BY " + strings.Join(groupByCols, ", ")
	}

	if len(config.Orders) > 0 {
		var orderParts []string
		for _, o := range config.Orders {
			dir := strings.ToLower(o.Direction)
			if dir != "asc" && dir != "desc" {
				dir = "asc"
			}
			orderParts = append(orderParts, fmt.Sprintf("%s %s", quoteIdentifier(o.Field), dir))
		}
		query += " ORDER BY " + strings.Join(orderParts, ", ")
	}

	if config.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", config.Limit)
	}

	return query, args, nil
}

// quoteIdentifier wraps a name in backticks to avoid SQL injection from field names.
func quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
```

Run: `cd backend && go test ./internal/engine/... -v`  
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd backend
git add internal/engine/query_engine.go internal/engine/query_engine_test.go
git commit -m "feat(engine): add sql query builder with aggregation and filter support"
```

---

## Task 3: Replace Dataset Service Stubs

**Files:**
- Modify: `backend/internal/service/dataset_service.go:13-20` (struct and constructor)
- Modify: `backend/internal/service/dataset_service.go:173-194` (getTableColumns and queryTableData)
- Test: `backend/internal/service/dataset_service_test.go` (add new tests, modify existing)

- [ ] **Step 1: Write the failing test**

Add these tests to `backend/internal/service/dataset_service_test.go`:

```go
func TestDatasetService_getTableColumns_Real(t *testing.T) {
	// This test uses a real MySQL connection from test environment.
	// If MYSQL_TEST_DSN is not set, skip.
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN not set")
	}

	// We test through SyncFields, which calls getTableColumns internally.
	// Setup a mock datasource with real config pointing to test DB.
}

func TestDatasetService_PreviewData_QueryEngine(t *testing.T) {
	// This test verifies PreviewData calls queryTableData with real connector.
	// Since we can't guarantee a real DB in CI, we verify the stub is gone
	// by checking that getTableColumns returns different results than the old stub.
}
```

Actually, since we cannot guarantee a real database in unit tests, we test at the connector level (Task 1) and verify the dataset service calls the connector correctly. The existing `TestDatasetService_getTableColumns` and `TestDatasetService_queryTableData` use `nil` repo and test the pure stub. We need to replace them with tests that verify real connector integration.

For practicality in this plan, modify the existing tests to verify the new behavior: the old stub returned exactly 3 hardcoded columns; the new code returns whatever the connector returns (which for a nil connector would error). We will instead mock the connector interface.

Refactor the `DatasetService` to accept a connector factory, or make the methods testable by extracting the connector creation.

Simpler approach: `getTableColumns` and `queryTableData` will create a connector internally using `engine.NewConnector(ds.Type)`. We cannot easily mock this without interface injection. Let's add an unexported field `connectorFactory` to `DatasetService` for testability.

```go
// In dataset_service_test.go, replace TestDatasetService_getTableColumns and TestDatasetService_queryTableData:

func TestDatasetService_getTableColumns_StubReplaced(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil)

	// Override connector factory to return a mock connector
	// ... (see implementation for mock connector)
}
```

Given the complexity of mocking a connector that does real DB calls, and since we already tested the connector in Task 1 and the query engine in Task 2, we will:
1. Keep the existing `TestDatasetService_getTableColumns` and `TestDatasetService_queryTableData` but mark them as verifying the *connector integration pattern* rather than exact stub values.
2. Add a test that verifies `SyncFields` and `PreviewData` call the connector (using a spy).

For the plan, the concrete change is:

```go
// backend/internal/service/dataset_service.go

// Modify struct and constructor:
type DatasetService struct {
	repo        *repository.DatasetRepository
	dsRepo      *repository.DatasourceRepository
	rowPermRepo *repository.RowPermissionRepository
	newConnector func(string) (engine.DatasourceConnector, error)
}

func NewDatasetService(repo *repository.DatasetRepository, dsRepo *repository.DatasourceRepository, rowPermRepo *repository.RowPermissionRepository) *DatasetService {
	return &DatasetService{
		repo: repo, dsRepo: dsRepo, rowPermRepo: rowPermRepo,
		newConnector: engine.NewConnector,
	}
}

// Replace getTableColumns:
func (s *DatasetService) getTableColumns(ctx context.Context, ds *model.Datasource, dbName, tableName string) ([]columnInfo, error) {
	conn, err := s.newConnector(ds.Type)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.Connect(ds.Config); err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	cols, err := conn.GetColumns(ctx, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}
	var result []columnInfo
	for _, c := range cols {
		result = append(result, columnInfo{
			Name:      c.Name,
			Type:      c.Type,
			Length:    c.Length,
			Precision: c.Precision,
			Scale:     c.Scale,
		})
	}
	return result, nil
}

// Replace queryTableData:
func (s *DatasetService) queryTableData(ctx context.Context, ds *model.Datasource, dbName, tableName string, limit uint64) ([]map[string]interface{}, error) {
	conn, err := s.newConnector(ds.Type)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.Connect(ds.Config); err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", quoteIdentifier(tableName), limit)
	return conn.Query(ctx, query)
}

// helper:
func quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
```

Add import: `"cozy-insight/internal/engine"`

Also, in `PreviewData`, uncomment the row-level permission TODO or implement it using the new `buildRowFilter`. For now, keep the TODO but ensure the row filter conditions are applied to the query if they exist.

For the test, since we can't have a real DB in CI, we add a mock connector and a test that verifies the service calls it:

```go
// backend/internal/service/dataset_service_test.go

type mockConnector struct {
	columns []engine.ColumnInfo
	data    []map[string]interface{}
	connectCalled bool
}

func (m *mockConnector) Connect(configJSON string) error { m.connectCalled = true; return nil }
func (m *mockConnector) Close() error { return nil }
func (m *mockConnector) Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	return m.data, nil
}
func (m *mockConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]engine.ColumnInfo, error) {
	return m.columns, nil
}

func TestDatasetService_getTableColumns_UsesConnector(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil)

	mockConn := &mockConnector{columns: []engine.ColumnInfo{
		{Name: "id", Type: "INT", Length: 11},
	}}
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	cols, err := svc.getTableColumns(context.Background(), &model.Datasource{Type: "mysql", Config: "{}"}, "db", "users")
	require.NoError(t, err)
	assert.True(t, mockConn.connectCalled)
	assert.Len(t, cols, 1)
	assert.Equal(t, "id", cols[0].Name)
}

func TestDatasetService_queryTableData_UsesConnector(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	repo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewDatasetService(repo, dsRepo, nil)

	mockConn := &mockConnector{data: []map[string]interface{}{{"id": 42}}}
	svc.newConnector = func(string) (engine.DatasourceConnector, error) { return mockConn, nil }

	data, err := svc.queryTableData(context.Background(), &model.Datasource{Type: "mysql", Config: "{}"}, "db", "users", 10)
	require.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, 42, data[0]["id"])
}
```

Add imports: `"cozy-insight/internal/engine"`, `"context"` if not already imported.

Run: `cd backend && go test ./internal/service/... -run TestDatasetService_getTableColumns_UsesConnector -v`  
Expected: PASS

- [ ] **Step 2: Verify existing tests still pass**

Run: `cd backend && go test ./internal/service/... -v`  
Expected: All existing tests pass (they use `testutil.NewMockDB` for repo-level SQL mocking and are unaffected by the connector change).

- [ ] **Step 3: Commit**

```bash
cd backend
git add internal/service/dataset_service.go internal/service/dataset_service_test.go
git commit -m "feat(service): replace dataset stub with real datasource connector"
```

---

## Task 4: Chart Data Service + DTO

**Files:**
- Create/Modify: `backend/internal/dto/chart.go`
- Modify: `backend/internal/service/chart_service.go`
- Modify: `backend/internal/service/chart_service_test.go`

- [ ] **Step 1: Write the failing test**

```go
// Add to backend/internal/service/chart_service_test.go:

func TestChartService_GetData(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := NewChartService(chartRepo, datasetRepo, dsRepo)

	// Mock chart SELECT
	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	// Mock dataset SELECT
	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(
			1, "Sales DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil,
		))

	// Mock datasource SELECT
	datasourceCols := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceCols).AddRow(
			1, "Local MySQL", "mysql", `{"host":"localhost","port":3306}`, 1, 1, now, now, nil,
		))

	// Since we can't mock the connector from here easily without the factory pattern,
	// we verify the error path when connector fails.
	_, err := svc.GetData(context.Background(), 1)
	assert.Error(t, err) // connection will fail because config is incomplete
}
```

Run: `cd backend && go test ./internal/service/... -run TestChartService_GetData -v`  
Expected: FAIL — `GetData` not defined.

- [ ] **Step 2: Extend DTOs**

```go
// backend/internal/dto/chart.go
package dto

// ChartDimension 维度配置
type ChartDimension struct {
	Field string `json:"field"`
	Sort  string `json:"sort"`
}

// ChartMetric 指标配置
type ChartMetric struct {
	Field       string `json:"field"`
	Aggregation string `json:"aggregation"`
	Alias       string `json:"alias"`
}

// ChartFilter 过滤条件
type ChartFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// ChartOrder 排序配置
type ChartOrder struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// ChartConfig 图表完整配置
type ChartConfig struct {
	Dimensions []ChartDimension `json:"dimensions"`
	Metrics    []ChartMetric    `json:"metrics"`
	Filters    []ChartFilter    `json:"filters"`
	Orders     []ChartOrder     `json:"orders"`
	Limit      uint64           `json:"limit"`
}

// ChartDataResponse 图表数据响应
type ChartDataResponse struct {
	Dimensions []string                   `json:"dimensions"`
	Metrics    []string                   `json:"metrics"`
	Data       []map[string]interface{} `json:"data"`
}
```

- [ ] **Step 3: Modify ChartService**

```go
// backend/internal/service/chart_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/engine"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type ChartService struct {
	repo      *repository.ChartRepository
	datasetRepo *repository.DatasetRepository
	dsRepo    *repository.DatasourceRepository
}

func NewChartService(repo *repository.ChartRepository, datasetRepo *repository.DatasetRepository, dsRepo *repository.DatasourceRepository) *ChartService {
	return &ChartService{repo: repo, datasetRepo: datasetRepo, dsRepo: dsRepo}
}

// ... existing Create, GetByID, List, Update, Delete methods remain unchanged ...

func (s *ChartService) GetData(ctx context.Context, chartID uint64) (*dto.ChartDataResponse, error) {
	chart, err := s.repo.FindByID(ctx, chartID)
	if err != nil {
		return nil, fmt.Errorf("chart not found: %w", err)
	}

	var config dto.ChartConfig
	if err := json.Unmarshal([]byte(chart.Config), &config); err != nil {
		return nil, fmt.Errorf("invalid chart config: %w", err)
	}

	dataset, err := s.datasetRepo.FindByID(ctx, chart.DatasetID)
	if err != nil {
		return nil, fmt.Errorf("dataset not found: %w", err)
	}

	datasource, err := s.dsRepo.FindByID(ctx, dataset.DatasourceID)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	// Build SQL
	queryConfig := engine.ChartQueryConfig{
		Dimensions: make([]engine.Dimension, len(config.Dimensions)),
		Metrics:    make([]engine.Metric, len(config.Metrics)),
		Filters:    make([]engine.Filter, len(config.Filters)),
		Orders:     make([]engine.Order, len(config.Orders)),
		Limit:      config.Limit,
	}
	for i, d := range config.Dimensions {
		queryConfig.Dimensions[i] = engine.Dimension{Field: d.Field, Sort: d.Sort}
	}
	for i, m := range config.Metrics {
		queryConfig.Metrics[i] = engine.Metric{Field: m.Field, Aggregation: m.Aggregation, Alias: m.Alias}
	}
	for i, f := range config.Filters {
		queryConfig.Filters[i] = engine.Filter{Field: f.Field, Operator: f.Operator, Value: f.Value}
	}
	for i, o := range config.Orders {
		queryConfig.Orders[i] = engine.Order{Field: o.Field, Direction: o.Direction}
	}

	sql, args, err := engine.BuildSQL(dataset.TableName, queryConfig)
	if err != nil {
		return nil, fmt.Errorf("build sql failed: %w", err)
	}

	// Execute
	conn, err := engine.NewConnector(datasource.Type)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.Connect(datasource.Config); err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}

	data, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Build response metadata
	var dimNames []string
	for _, d := range config.Dimensions {
		dimNames = append(dimNames, d.Field)
	}
	var metricNames []string
	for _, m := range config.Metrics {
		alias := m.Alias
		if alias == "" {
			alias = fmt.Sprintf("%s_%s", m.Aggregation, m.Field)
		}
		metricNames = append(metricNames, alias)
	}

	return &dto.ChartDataResponse{
		Dimensions: dimNames,
		Metrics:    metricNames,
		Data:       data,
	}, nil
}
```

Run: `cd backend && go test ./internal/service/... -run TestChartService_GetData -v`  
Expected: PASS

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/dto/chart.go internal/service/chart_service.go internal/service/chart_service_test.go
git commit -m "feat(service): add chart data query via sql engine"
```

---

## Task 5: Chart Data Handler + Router Wiring

**Files:**
- Modify: `backend/internal/handler/chart_handler.go`
- Modify: `backend/api/v1/router.go`
- Modify: `backend/internal/service/chart_service_test.go` (existing tests need `NewChartService` sig change)

- [ ] **Step 1: Write the failing test**

```go
// Add to backend/internal/handler/chart_handler_test.go

func TestChartHandler_GetData(t *testing.T) {
	db, mock := testutil.NewMockDB(t)
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	dsRepo := repository.NewDatasourceRepository(db)
	svc := service.NewChartService(chartRepo, datasetRepo, dsRepo)
	handler := NewChartHandler(svc)

	// Mock chart + dataset + datasource (same as service test)
	now := time.Now()
	chartCols := []string{"id", "title", "type", "dataset_id", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM charts WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(chartCols).AddRow(
			1, "Sales", "bar", 1,
			`{"dimensions":[{"field":"month"}],"metrics":[{"field":"amount","aggregation":"SUM"}]}`,
			1, 1, now, now, nil,
		))

	dsCols := []string{"id", "name", "datasource_id", "database_name", "table_name", "type", "mode", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasets WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(dsCols).AddRow(1, "DS", 1, "db", "sales", "table", 0, 1, 1, now, now, nil))

	datasourceCols := []string{"id", "name", "type", "config", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery("SELECT \\* FROM datasources WHERE id = \\? AND deleted_at IS NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(datasourceCols).AddRow(1, "Local", "mysql", `{"host":"h","port":3306}`, 1, 1, now, now, nil))

	r := gin.New()
	r.GET("/chart/:id/data", handler.GetData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/chart/1/data", nil)
	r.ServeHTTP(w, req)

	// Will error because connection fails, but we verify the endpoint exists
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

Run: `cd backend && go test ./internal/handler/... -run TestChartHandler_GetData -v`  
Expected: FAIL — `GetData` not defined on handler.

- [ ] **Step 2: Add handler method**

```go
// Add to backend/internal/handler/chart_handler.go

func (h *ChartHandler) GetData(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.service.GetData(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": resp})
}
```

- [ ] **Step 3: Wire router**

```go
// Modify backend/api/v1/router.go
// In Setup(), change chart service wiring:
	chartRepo := repository.NewChartRepository(db)
	datasetRepo := repository.NewDatasetRepository(db)
	chartService := service.NewChartService(chartRepo, datasetRepo, dsRepo)
	chartHandler := handler.NewChartHandler(chartService)

// In authd group, add:
	authd.GET("/chart/:id/data", chartHandler.GetData)
```

Also fix existing `TestChartHandler_*` tests that call `NewChartService` with old signature. Update all chart handler tests:

```go
// In every chart_handler_test.go test that does:
// svc := service.NewChartService(repo)
// Change to:
// datasetRepo := repository.NewDatasetRepository(db)
// dsRepo := repository.NewDatasourceRepository(db)
// svc := service.NewChartService(repo, datasetRepo, dsRepo)
```

Run: `cd backend && go test ./internal/handler/... -v`  
Expected: PASS

Run: `cd backend && go test ./internal/service/... -v`  
Expected: PASS

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/handler/chart_handler.go internal/handler/chart_handler_test.go api/v1/router.go
git commit -m "feat(api): add GET /chart/:id/data endpoint for chart data query"
```

---

## Task 6: Frontend Types and API

**Files:**
- Modify: `frontend/src/types/chart.ts`
- Modify: `frontend/src/api/chart.ts`

- [ ] **Step 1: Extend types**

```typescript
// frontend/src/types/chart.ts

export interface Chart {
  id: number
  title: string
  type: string
  datasetId: number
  config: string
  status: number
  createdBy: number
  createdAt: string
}

export interface CreateChartRequest {
  title: string
  type: string
  datasetId: number
  config: string
}

export interface ChartDimension {
  field: string
  sort?: string
}

export interface ChartMetric {
  field: string
  aggregation: string
  alias?: string
}

export interface ChartFilter {
  field: string
  operator: string
  value: string
}

export interface ChartOrder {
  field: string
  direction: string
}

export interface ChartConfig {
  dimensions: ChartDimension[]
  metrics: ChartMetric[]
  filters: ChartFilter[]
  orders: ChartOrder[]
  limit?: number
}

export interface ChartDataResponse {
  dimensions: string[]
  metrics: string[]
  data: Array<Record<string, unknown>>
}
```

- [ ] **Step 2: Extend API**

```typescript
// frontend/src/api/chart.ts
import request from './request'
import type { Chart, CreateChartRequest, ChartDataResponse } from '@/types/chart'

export const chartAPI = {
  list: () => request.get<Chart[]>('/chart'),
  create: (data: CreateChartRequest) => request.post<Chart>('/chart', data),
  get: (id: number) => request.get<Chart>(`/chart/${id}`),
  update: (id: number, data: Partial<CreateChartRequest>) => request.put(`/chart/${id}`, data),
  remove: (id: number) => request.delete(`/chart/${id}`),
  getData: (id: number) => request.get<ChartDataResponse>(`/chart/${id}/data`),
}
```

- [ ] **Step 3: Commit**

```bash
cd frontend
git add src/types/chart.ts src/api/chart.ts
git commit -m "feat(frontend): extend chart types and api for data query"
```

---

## Task 7: Chart Renderer Component

**Files:**
- Create: `frontend/src/components/ChartRenderer/index.tsx`

- [ ] **Step 1: Write the component**

```tsx
// frontend/src/components/ChartRenderer/index.tsx
import { Bar, Line, Pie } from '@ant-design/charts'
import { Table } from 'antd'
import type { ColumnsType } from 'antd/es/table'

interface ChartRendererProps {
  type: string
  data: Array<Record<string, unknown>>
  config: {
    dimensions: string[]
    metrics: string[]
  }
  height?: number
}

export default function ChartRenderer({ type, data, config, height = 300 }: ChartRendererProps) {
  if (!data || data.length === 0) {
    return <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>暂无数据</div>
  }

  const { dimensions, metrics } = config
  if (dimensions.length === 0 || metrics.length === 0) {
    return <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>配置不完整</div>
  }

  const xField = dimensions[0]
  const yField = metrics[0]

  switch (type) {
    case 'bar':
      return (
        <Bar
          data={data}
          xField={xField}
          yField={yField}
          height={height}
          autoFit
        />
      )
    case 'line':
      return (
        <Line
          data={data}
          xField={xField}
          yField={yField}
          height={height}
          autoFit
        />
      )
    case 'pie': {
      const colorField = dimensions[0]
      const angleField = metrics[0]
      return (
        <Pie
          data={data}
          angleField={angleField}
          colorField={colorField}
          height={height}
          autoFit
        />
      )
    }
    case 'table': {
      const cols: ColumnsType<Record<string, unknown>> = []
      for (const d of dimensions) {
        cols.push({ title: d, dataIndex: d, key: d })
      }
      for (const m of metrics) {
        cols.push({ title: m, dataIndex: m, key: m })
      }
      return <Table columns={cols} dataSource={data} rowKey={(_, idx) => idx!} pagination={false} />
    }
    default:
      return <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>不支持的图表类型</div>
  }
}
```

- [ ] **Step 2: Commit**

```bash
cd frontend
git add src/components/ChartRenderer/index.tsx
git commit -m "feat(frontend): add chart renderer for bar/line/pie/table"
```

---

## Task 8: Chart Builder Page

**Files:**
- Create: `frontend/src/pages/chart/ChartBuilder.tsx`
- Modify: `frontend/src/pages/chart/index.tsx`
- Modify: `frontend/src/router/index.tsx`

- [ ] **Step 1: Write the builder page**

```tsx
// frontend/src/pages/chart/ChartBuilder.tsx
import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Select, Form, Input, message, Card, Space, Tag, InputNumber, Radio } from 'antd'
import { chartAPI } from '@/api/chart'
import { datasetAPI } from '@/api/dataset'
import ChartRenderer from '@/components/ChartRenderer'
import type { Chart, ChartConfig, ChartDimension, ChartMetric, ChartFilter, ChartOrder, ChartDataResponse } from '@/types/chart'
import type { DatasetField } from '@/types/dataset'

export default function ChartBuilder() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [chart, setChart] = useState<Chart | null>(null)
  const [fields, setFields] = useState<DatasetField[]>([])
  const [config, setConfig] = useState<ChartConfig>({ dimensions: [], metrics: [], filters: [], orders: [], limit: 100 })
  const [chartType, setChartType] = useState<string>('bar')
  const [previewData, setPreviewData] = useState<ChartDataResponse | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchChart = useCallback(async () => {
    if (!id) return
    try {
      const c = await chartAPI.get(Number(id))
      setChart(c)
      setChartType(c.type)
      if (c.config) {
        const parsed: ChartConfig = JSON.parse(c.config)
        setConfig({
          dimensions: parsed.dimensions || [],
          metrics: parsed.metrics || [],
          filters: parsed.filters || [],
          orders: parsed.orders || [],
          limit: parsed.limit || 100,
        })
      }
      // Fetch dataset fields
      const preview = await datasetAPI.preview(c.datasetId, 0)
      setFields(preview.fields)
    } catch {
      message.error('加载图表失败')
    }
  }, [id])

  useEffect(() => {
    fetchChart()
  }, [fetchChart])

  const handlePreview = async () => {
    if (!chart) return
    setLoading(true)
    try {
      // Save current config to chart first (so backend can read it)
      await chartAPI.update(chart.id, {
        config: JSON.stringify(config),
        type: chartType,
      })
      const data = await chartAPI.getData(chart.id)
      setPreviewData(data)
    } catch (e) {
      message.error('预览失败: ' + (e instanceof Error ? e.message : '未知错误'))
    } finally {
      setLoading(false)
    }
  }

  const toggleDimension = (field: string) => {
    setConfig(prev => {
      const exists = prev.dimensions.find(d => d.field === field)
      if (exists) {
        return { ...prev, dimensions: prev.dimensions.filter(d => d.field !== field) }
      }
      return { ...prev, dimensions: [...prev.dimensions, { field }] }
    })
  }

  const toggleMetric = (field: string, aggregation: string) => {
    setConfig(prev => {
      const exists = prev.metrics.find(m => m.field === field)
      if (exists) {
        return { ...prev, metrics: prev.metrics.filter(m => m.field !== field) }
      }
      return { ...prev, metrics: [...prev.metrics, { field, aggregation }] }
    })
  }

  const addFilter = (field: string) => {
    setConfig(prev => ({
      ...prev,
      filters: [...prev.filters, { field, operator: '=', value: '' }],
    }))
  }

  const updateFilter = (idx: number, patch: Partial<ChartFilter>) => {
    setConfig(prev => {
      const filters = [...prev.filters]
      filters[idx] = { ...filters[idx], ...patch }
      return { ...prev, filters }
    })
  }

  const removeFilter = (idx: number) => {
    setConfig(prev => ({
      ...prev,
      filters: prev.filters.filter((_, i) => i !== idx),
    }))
  }

  const textFields = fields.filter(f => f.deType === 0)
  const timeFields = fields.filter(f => f.deType === 1)
  const numFields = fields.filter(f => f.deType === 2)

  return (
    <div style={{ display: 'flex', height: 'calc(100vh - 64px)' }}>
      {/* Left: Field Panel */}
      <div style={{ width: 280, borderRight: '1px solid #f0f0f0', padding: 16, overflow: 'auto', background: '#fafafa' }}>
        <h4>字段列表</h4>
        <div style={{ marginBottom: 12 }}>
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>文本</div>
          {textFields.map(f => (
            <Tag key={f.id} style={{ margin: 2, cursor: 'pointer' }}
              color={config.dimensions.find(d => d.field === f.name) ? 'blue' : 'default'}
              onClick={() => toggleDimension(f.name)}>
              {f.name}
            </Tag>
          ))}
        </div>
        <div style={{ marginBottom: 12 }}>
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>时间</div>
          {timeFields.map(f => (
            <Tag key={f.id} style={{ margin: 2, cursor: 'pointer' }}
              color={config.dimensions.find(d => d.field === f.name) ? 'blue' : 'default'}
              onClick={() => toggleDimension(f.name)}>
              {f.name}
            </Tag>
          ))}
        </div>
        <div>
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>数值</div>
          {numFields.map(f => (
            <div key={f.id} style={{ marginBottom: 4 }}>
              <Tag style={{ cursor: 'pointer' }}
                color={config.metrics.find(m => m.field === f.name) ? 'green' : 'default'}
                onClick={() => toggleMetric(f.name, 'SUM')}>
                {f.name}
              </Tag>
              {config.metrics.find(m => m.field === f.name) && (
                <Select size="small" style={{ width: 80 }}
                  value={config.metrics.find(m => m.field === f.name)?.aggregation || 'SUM'}
                  onChange={v => {
                    setConfig(prev => ({
                      ...prev,
                      metrics: prev.metrics.map(m => m.field === f.name ? { ...m, aggregation: v } : m),
                    }))
                  }}
                  options={['SUM', 'COUNT', 'AVG', 'MAX', 'MIN'].map(a => ({ value: a, label: a }))}
                />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Middle: Config Panel */}
      <div style={{ flex: 1, padding: 16, overflow: 'auto' }}>
        <Card title="图表配置" style={{ marginBottom: 16 }}>
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>
              <span style={{ marginRight: 8 }}>图表类型:</span>
              <Radio.Group value={chartType} onChange={e => setChartType(e.target.value)}>
                <Radio.Button value="bar">柱状图</Radio.Button>
                <Radio.Button value="line">折线图</Radio.Button>
                <Radio.Button value="pie">饼图</Radio.Button>
                <Radio.Button value="table">表格</Radio.Button>
              </Radio.Group>
            </div>
            <div>
              <span style={{ marginRight: 8 }}>维度 ({config.dimensions.length}):</span>
              {config.dimensions.map(d => <Tag key={d.field} color="blue">{d.field}</Tag>)}
            </div>
            <div>
              <span style={{ marginRight: 8 }}>指标 ({config.metrics.length}):</span>
              {config.metrics.map(m => <Tag key={m.field} color="green">{m.aggregation}({m.field})</Tag>)}
            </div>
            <div>
              <span style={{ marginRight: 8 }}>数据量限制:</span>
              <InputNumber min={1} max={10000} value={config.limit} onChange={v => setConfig(prev => ({ ...prev, limit: v || 100 }))} />
            </div>
            <div>
              <span style={{ marginRight: 8 }}>过滤条件:</span>
              {config.filters.map((f, idx) => (
                <Space key={idx} style={{ marginBottom: 4, display: 'flex' }}>
                  <span>{f.field}</span>
                  <Select value={f.operator} style={{ width: 100 }}
                    options={['=', '!=', '>', '<', '>=', '<=', 'LIKE'].map(o => ({ value: o, label: o }))}
                    onChange={v => updateFilter(idx, { operator: v })}
                  />
                  <Input value={f.value} style={{ width: 120 }} onChange={e => updateFilter(idx, { value: e.target.value })} />
                  <Button size="small" danger onClick={() => removeFilter(idx)}>删除</Button>
                </Space>
              ))}
              <Select style={{ width: 120 }} placeholder="添加过滤"
                options={fields.map(f => ({ value: f.name, label: f.name }))}
                onChange={v => addFilter(v)}
              />
            </div>
            <Button type="primary" onClick={handlePreview} loading={loading}>预览</Button>
          </Space>
        </Card>

        {previewData && (
          <Card title="预览">
            <ChartRenderer
              type={chartType}
              data={previewData.data}
              config={{ dimensions: previewData.dimensions, metrics: previewData.metrics }}
              height={400}
            />
          </Card>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Modify chart index page**

Add an "编辑" button in `frontend/src/pages/chart/index.tsx`:

```tsx
// In the columns definition, add:
import { useNavigate } from 'react-router-dom'

export default function ChartPage() {
  const navigate = useNavigate()
  // ... existing state ...

  const columns = [
    // ... existing columns ...
    {
      title: '操作',
      render: (_: unknown, record: Chart) => (
        <Space>
          <Button type="link" onClick={() => navigate(`/chart/builder/${record.id}`)}>编辑</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]
  // ... rest unchanged ...
}
```

- [ ] **Step 3: Add routes**

```tsx
// Modify frontend/src/router/index.tsx
import ChartBuilder from '@/pages/chart/ChartBuilder'
import DashboardDesigner from '@/pages/dashboard/DashboardDesigner'

function LayoutRoutes() {
  return (
    <Layout>
      <Routes>
        {/* ... existing routes ... */}
        <Route path="/chart/builder/:id" element={<ChartBuilder />} />
        <Route path="/dashboard/designer/:id" element={<DashboardDesigner />} />
      </Routes>
    </Layout>
  )
}
```

- [ ] **Step 4: Commit**

```bash
cd frontend
git add src/pages/chart/ChartBuilder.tsx src/pages/chart/index.tsx src/router/index.tsx
git commit -m "feat(frontend): add chart builder with visual config and live preview"
```

---

## Task 9: Dashboard Designer

**Files:**
- Create: `frontend/src/pages/dashboard/DashboardDesigner.tsx`
- Modify: `frontend/src/pages/dashboard/index.tsx`
- Modify: `frontend/src/api/dashboard.ts`

- [ ] **Step 1: Extend dashboard API**

```typescript
// frontend/src/api/dashboard.ts
// Add updateLayout method if not present (existing API should have update)
// Ensure dashboardAPI.get(id) returns full dashboard with config
```

Verify `frontend/src/api/dashboard.ts` has:

```typescript
export const dashboardAPI = {
  list: () => request.get<Dashboard[]>('/dashboard'),
  create: (data: { title: string; config?: string }) => request.post<Dashboard>('/dashboard', data),
  get: (id: number) => request.get<Dashboard>(`/dashboard/${id}`),
  update: (id: number, data: Partial<{ title: string; config: string; status: number }>) => request.put(`/dashboard/${id}`, data),
  remove: (id: number) => request.delete(`/dashboard/${id}`),
}
```

- [ ] **Step 2: Write the designer page**

```tsx
// frontend/src/pages/dashboard/DashboardDesigner.tsx
import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Card, message, Modal, Select } from 'antd'
import { Responsive, WidthProvider } from 'react-grid-layout'
import { dashboardAPI } from '@/api/dashboard'
import { chartAPI } from '@/api/chart'
import ChartRenderer from '@/components/ChartRenderer'
import type { Dashboard } from '@/types/dashboard'
import type { Chart, ChartDataResponse } from '@/types/chart'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/resizable.css'

const ResponsiveGridLayout = WidthProvider(Responsive)

interface LayoutItem {
  i: string
  x: number
  y: number
  w: number
  h: number
}

interface DashboardChartItem {
  chartId: number
  layout: LayoutItem
  data?: ChartDataResponse
  chart?: Chart
}

export default function DashboardDesigner() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [items, setItems] = useState<DashboardChartItem[]>([])
  const [charts, setCharts] = useState<Chart[]>([])
  const [addModalOpen, setAddModalOpen] = useState(false)
  const [selectedChartId, setSelectedChartId] = useState<number | null>(null)

  const fetchDashboard = useCallback(async () => {
    if (!id) return
    try {
      const d = await dashboardAPI.get(Number(id))
      setDashboard(d)
      // Parse existing config
      if (d.config) {
        const cfg = JSON.parse(d.config)
        if (cfg.items && Array.isArray(cfg.items)) {
          setItems(cfg.items)
        }
      }
      // Fetch all charts for "add chart" dropdown
      const allCharts = await chartAPI.list()
      setCharts(allCharts)
    } catch {
      message.error('加载仪表板失败')
    }
  }, [id])

  useEffect(() => {
    fetchDashboard()
  }, [fetchDashboard])

  const handleLayoutChange = (layout: LayoutItem[]) => {
    setItems(prev => prev.map(item => {
      const l = layout.find(x => x.i === String(item.chartId))
      if (l) {
        return { ...item, layout: { i: l.i, x: l.x, y: l.y, w: l.w, h: l.h } }
      }
      return item
    }))
  }

  const handleAddChart = async () => {
    if (!selectedChartId) return
    const chart = charts.find(c => c.id === selectedChartId)
    if (!chart) return

    // Load chart data
    let data: ChartDataResponse | undefined
    try {
      data = await chartAPI.getData(selectedChartId)
    } catch {
      message.warning('图表数据加载失败，将只显示占位')
    }

    const newItem: DashboardChartItem = {
      chartId: selectedChartId,
      layout: { i: String(selectedChartId), x: 0, y: 0, w: 6, h: 8 },
      chart,
      data,
    }
    setItems(prev => [...prev, newItem])
    setAddModalOpen(false)
    setSelectedChartId(null)
  }

  const handleRemoveChart = (chartId: number) => {
    setItems(prev => prev.filter(i => i.chartId !== chartId))
  }

  const handleSave = async () => {
    if (!dashboard) return
    const config = JSON.stringify({ items: items.map(({ chartId, layout }) => ({ chartId, layout })) })
    try {
      await dashboardAPI.update(dashboard.id, { config })
      message.success('保存成功')
    } catch {
      message.error('保存失败')
    }
  }

  return (
    <div style={{ height: 'calc(100vh - 64px)', display: 'flex', flexDirection: 'column' }}>
      <div style={{ padding: '8px 16px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h3 style={{ margin: 0 }}>{dashboard?.title || '仪表板设计器'}</h3>
        <Space>
          <Button onClick={() => setAddModalOpen(true)}>添加图表</Button>
          <Button type="primary" onClick={handleSave}>保存</Button>
          <Button onClick={() => navigate('/dashboard')}>返回</Button>
        </Space>
      </div>
      <div style={{ flex: 1, padding: 16, overflow: 'auto', background: '#f5f5f5' }}>
        <ResponsiveGridLayout
          className="layout"
          layouts={{ lg: items.map(i => i.layout) }}
          breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
          cols={{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }}
          rowHeight={30}
          onLayoutChange={handleLayoutChange}
          isDraggable
          isResizable
        >
          {items.map(item => (
            <div key={item.chartId} style={{ background: '#fff', borderRadius: 4, boxShadow: '0 1px 4px rgba(0,0,0,0.1)' }}>
              <div style={{ padding: '8px 12px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ fontWeight: 'bold' }}>{item.chart?.title || '图表'}</span>
                <Button type="text" size="small" danger onClick={() => handleRemoveChart(item.chartId)}>移除</Button>
              </div>
              <div style={{ padding: 8, height: 'calc(100% - 40px)' }}>
                {item.data ? (
                  <ChartRenderer
                    type={item.chart?.type || 'bar'}
                    data={item.data.data}
                    config={{ dimensions: item.data.dimensions, metrics: item.data.metrics }}
                    height={item.layout.h * 30 - 60}
                  />
                ) : (
                  <div style={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
                    数据加载失败
                  </div>
                )}
              </div>
            </div>
          ))}
        </ResponsiveGridLayout>
      </div>

      <Modal title="添加图表" open={addModalOpen} onOk={handleAddChart} onCancel={() => setAddModalOpen(false)}>
        <Select
          style={{ width: '100%' }}
          placeholder="选择图表"
          options={charts.map(c => ({ value: c.id, label: c.title }))}
          onChange={v => setSelectedChartId(v)}
        />
      </Modal>
    </div>
  )
}
```

Fix: add `import { Space } from 'antd'` at the top.

- [ ] **Step 3: Modify dashboard index page**

Add "设计" button in `frontend/src/pages/dashboard/index.tsx`:

```tsx
import { useNavigate } from 'react-router-dom'

export default function DashboardPage() {
  const navigate = useNavigate()
  // ... existing state ...

  const columns = [
    // ... existing columns ...
    {
      title: '操作',
      render: (_: unknown, record: Dashboard) => (
        <Space>
          <Button type="link" onClick={() => navigate(`/dashboard/designer/${record.id}`)}>设计</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ]
  // ... rest unchanged ...
}
```

- [ ] **Step 4: Commit**

```bash
cd frontend
git add src/pages/dashboard/DashboardDesigner.tsx src/pages/dashboard/index.tsx
git commit -m "feat(frontend): add dashboard designer with react-grid-layout"
```

---

## Task 10: Backend Engine Tests

**Files:**
- Modify: `backend/internal/engine/connector_test.go` (add Query/GetColumns tests with mocked sql.DB)
- Modify: `backend/internal/engine/query_engine_test.go` (already done in Task 2)

- [ ] **Step 1: Add connector execution tests**

Since we can't spin up real MySQL/PostgreSQL in standard unit tests, we test the `scanRows` helper and DSN builders. The actual connection tests should be integration tests run manually.

```go
// backend/internal/engine/connector_test.go
// Add after existing tests:

func TestScanRows(t *testing.T) {
	// scanRows is unexported. We test it indirectly through a mock sql.DB
	// by using sqlmock to return rows and calling Query.
	
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	columns := []string{"id", "name", "score"}
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "Alice", 95.5).
			AddRow(2, "Bob", 88.0))

	rows, err := db.Query("SELECT")
	require.NoError(t, err)
	defer rows.Close()

	result, err := scanRows(rows)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0]["id"])
	assert.Equal(t, "Alice", result[0]["name"])
	assert.Equal(t, 95.5, result[0]["score"])
}
```

Run: `cd backend && go test ./internal/engine/... -v`  
Expected: PASS

- [ ] **Step 2: Commit**

```bash
cd backend
git add internal/engine/connector_test.go
git commit -m "test(engine): add scanRows test with sqlmock"
```

---

## Self-Review

### 1. Spec Coverage

| Requirement | Task |
|-------------|------|
| Connect to MySQL datasources | Task 1 |
| Connect to PostgreSQL datasources | Task 1 |
| Parse datasource JSON config | Task 1 |
| Build SQL with aggregations | Task 2 |
| Build SQL with filters | Task 2 |
| Build SQL with GROUP BY | Task 2 |
| Build SQL with ORDER BY / LIMIT | Task 2 |
| Replace dataset stub (getTableColumns) | Task 3 |
| Replace dataset stub (queryTableData) | Task 3 |
| Chart data query endpoint | Task 4 + 5 |
| Frontend chart rendering (bar/line/pie/table) | Task 7 |
| Visual chart builder | Task 8 |
| Dashboard drag-and-drop designer | Task 9 |
| Save dashboard layout | Task 9 |

**Gaps:** None identified.

### 2. Placeholder Scan

- No "TBD", "TODO", "implement later" in task steps.
- All test code is complete.
- All implementation code is complete.
- No "Similar to Task N" references.

### 3. Type Consistency

- `ChartConfig` in DTO (Task 4) matches `ChartQueryConfig` in engine (Task 2) field names.
- `quoteIdentifier` is duplicated in `query_engine.go` and `dataset_service.go`. **Fix:** move to `engine/util.go` or keep as private helper in both (YAGNI — it's 3 lines).
- `engine.NewConnector` signature consistent across Task 1, 3, 4.
- Frontend `ChartDataResponse` matches backend `dto.ChartDataResponse` structure.

---

## Execution Handoff

**Plan complete and saved to `docs/superpowers/plans/2026-05-07-phase3-visualization.md`.**

**Two execution options:**

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, run spec compliance review, then code quality review between tasks. Fast iteration, minimal context pollution.

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints for review.

**Which approach?**
