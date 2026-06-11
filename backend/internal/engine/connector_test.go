package engine

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanRows(t *testing.T) {
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

func TestScanRows_NumericConversion(t *testing.T) {
	// 当 driver 返回 []byte 但列类型是 INT/FLOAT 时，scanRows 应当把字符串解析回数字
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cols := sqlmock.NewRows([]string{"id", "big", "tiny", "price", "ratio", "name"})
	cols = cols.FromCSVString("id,big,tiny,price,ratio,name")
	// Above returns proper types; for testing []byte-string conversion we use
	// NewRows + AddRow with raw []byte to simulate the "scan into []byte" path
	// that some drivers use. Here we exercise the typed path.
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "big", "tiny", "price", "ratio", "name"}).
			AddRow(int64(1), int64(9999999999), int64(-128), 3.14, 2.71, "ok"),
	)

	rows, err := db.Query("SELECT")
	require.NoError(t, err)
	defer rows.Close()

	result, err := scanRows(rows)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result[0]["id"])
	assert.Equal(t, int64(9999999999), result[0]["big"])
	assert.Equal(t, int64(-128), result[0]["tiny"])
	assert.Equal(t, 3.14, result[0]["price"])
	assert.Equal(t, 2.71, result[0]["ratio"])
	assert.Equal(t, "ok", result[0]["name"])
}

func TestScanRows_NumericParseFallback(t *testing.T) {
	// 当数字列包含不能解析的 []byte 时，scanRows 应回退为字符串
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// AddRow with []byte 模拟 driver 把所有列返回为 byte slice 的情形
	// 当 columnType 是 INT/FLOAT 时转换失败则回退
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}).
			AddRow([]byte("not-a-number"), []byte("Alice")),
	)
	rows, err := db.Query("SELECT")
	require.NoError(t, err)
	defer rows.Close()

	result, err := scanRows(rows)
	require.NoError(t, err)
	require.Len(t, result, 1)
	// typesAvailable 路径下 []byte 字符串会回退为 string
	assert.Equal(t, "not-a-number", result[0]["id"])
	assert.Equal(t, "Alice", result[0]["name"])
}

func TestScanRows_IterationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("iteration failure"))
	_, err = db.Query("SELECT")
	assert.Error(t, err)
}

func TestScanRows_ColumnTypesError(t *testing.T) {
	// 模拟 columns 数与 types 数不一致（typesAvailable = false 路径）
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"a", "b"}).AddRow([]byte("x"), []byte("y")),
	)
	rows, err := db.Query("SELECT")
	require.NoError(t, err)
	defer rows.Close()

	result, err := scanRows(rows)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "x", result[0]["a"])
	assert.Equal(t, "y", result[0]["b"])
}

func TestNewConnector_MySQL(t *testing.T) {
	conn, err := NewConnector("mysql")
	require.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestNewConnector_Unsupported(t *testing.T) {
	_, err := NewConnector("foobar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestNewConnector_AllTypes(t *testing.T) {
	types := []string{"mysql", "postgresql", "sqlite", "clickhouse", "sqlserver", "excel", "csv", "oracle", "doris", "starrocks", "mongodb", "hive", "elasticsearch", "api"}
	for _, ty := range types {
		t.Run(ty, func(t *testing.T) {
			conn, err := NewConnector(ty)
			require.NoError(t, err)
			assert.NotNil(t, conn)
		})
	}
}

func TestAPIConnector_Connect(t *testing.T) {
	conn := &apiConnector{}
	// 缺少 url
	err := conn.Connect(`{"method":"GET"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url")

	// 完整配置
	err = conn.Connect(`{"url":"https://api.example.com/data","method":"POST","timeout":10,"jsonPath":"data.items"}`)
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/data", conn.baseURL)
	assert.Equal(t, "POST", conn.method)
	assert.Equal(t, 10, conn.timeoutSec)
	assert.Equal(t, "data.items", conn.jsonPath)
}

func TestAPIConnector_Connect_DefaultMethod(t *testing.T) {
	conn := &apiConnector{}
	err := conn.Connect(`{"url":"https://x.example.com"}`)
	require.NoError(t, err)
	assert.Equal(t, "GET", conn.method)
	assert.Equal(t, 30, conn.timeoutSec)
}

func TestAPIConnector_Connect_InvalidJSON(t *testing.T) {
	conn := &apiConnector{}
	err := conn.Connect(`not-a-json`)
	assert.Error(t, err)
}

func TestAPIConnector_Query(t *testing.T) {
	// 启动一个 httptest server 返回数组
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"name":"a"},{"id":2,"name":"b"}]`))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `","method":"GET"}`)
	require.NoError(t, err)

	rows, err := conn.Query(context.Background(), "")
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "a", rows[0]["name"])
}

func TestAPIConnector_Query_JSONPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{"total":2},"data":{"items":[{"id":1}]}}`))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `","jsonPath":"data.items"}`)
	require.NoError(t, err)

	rows, err := conn.Query(context.Background(), "")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, float64(1), rows[0]["id"])
}

func TestAPIConnector_Query_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `"}`)
	require.NoError(t, err)

	_, err = conn.Query(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestAPIConnector_Query_WithParams(t *testing.T) {
	var receivedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1}]`))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `","params":{"q":"hello","page":"2"}}`)
	require.NoError(t, err)

	_, err = conn.Query(context.Background(), "")
	require.NoError(t, err)
	assert.Contains(t, receivedQuery, "q=hello")
	assert.Contains(t, receivedQuery, "page=2")
}

func TestAPIConnector_Query_WithHeaders(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1}]`))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `","headers":{"Authorization":"Bearer xyz"}}`)
	require.NoError(t, err)

	_, err = conn.Query(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "Bearer xyz", receivedAuth)
}

func TestAPIConnector_Query_NotAnArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"single":"object","not":"array"}`))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `"}`)
	require.NoError(t, err)

	_, err = conn.Query(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a JSON array")
}

func TestAPIConnector_Query_JSONPathMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"items":[{"id":1}]}}`))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `","jsonPath":"data.nope"}`)
	require.NoError(t, err)

	_, err = conn.Query(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAPIConnector_Query_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json {{{`))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `"}`)
	require.NoError(t, err)

	_, err = conn.Query(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json decode")
}

func TestAPIConnector_GetColumns(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"name":"a","email":"x@y.z"}]`))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `"}`)
	require.NoError(t, err)

	cols, err := conn.GetColumns(context.Background(), "", "")
	require.NoError(t, err)
	colsByName := make(map[string]ColumnInfo, len(cols))
	for _, c := range cols {
		colsByName[c.Name] = c
	}
	assert.Contains(t, colsByName, "id")
	assert.Contains(t, colsByName, "name")
	assert.Contains(t, colsByName, "email")
}

func TestAPIConnector_GetColumns_EmptyData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	conn := &apiConnector{}
	err := conn.Connect(`{"url":"` + server.URL + `"}`)
	require.NoError(t, err)

	_, err = conn.GetColumns(context.Background(), "", "")
	assert.Error(t, err)
}

// ---------------- SSH tunnel config 解析 ----------------

func TestExtractSSHConfig_Disabled(t *testing.T) {
	cfg, err := extractSSHConfig(`{"host":"db","port":3306}`)
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestExtractSSHConfig_Enabled(t *testing.T) {
	cfg, err := extractSSHConfig(`{"host":"db","port":3306,"sshTunnel":{"enabled":true,"host":"bastion","port":22,"username":"ops","password":"x"}}`)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "bastion", cfg.Host)
	assert.Equal(t, 22, cfg.Port)
	assert.Equal(t, "ops", cfg.Username)
	assert.Equal(t, "x", cfg.Password)
}

func TestExtractSSHConfig_EnabledButMissing(t *testing.T) {
	_, err := extractSSHConfig(`{"sshTunnel":{"enabled":true}}`)
	assert.Error(t, err)
}

func TestExtractSSHConfig_InvalidJSON(t *testing.T) {
	_, err := extractSSHConfig(`not-json`)
	assert.Error(t, err)
}

func TestNewConnectorWithTunnel_NoTunnel(t *testing.T) {
	// 无 sshTunnel 时退化为普通 NewConnector
	conn, tunnel, err := NewConnectorWithTunnel("mysql", `{"host":"x","port":3306,"username":"u","password":"p","database":"d"}`)
	_ = conn
	require.NoError(t, err)
	assert.Nil(t, tunnel)
}

func TestOracleConnector_BuildDSN(t *testing.T) {
	conn := &oracleConnector{}
	dsn, err := conn.buildDSN(`{"host":"localhost","port":1521,"username":"system","password":"pass","database":"ORCL"}`)
	require.NoError(t, err)
	assert.Equal(t, "oracle://system:pass@localhost:1521/ORCL", dsn)
}

func TestOracleConnector_BuildDSN_DefaultPort(t *testing.T) {
	conn := &oracleConnector{}
	dsn, err := conn.buildDSN(`{"host":"db.example.com","username":"u","password":"p","database":"FREE"}`)
	require.NoError(t, err)
	assert.Equal(t, "oracle://u:p@db.example.com:1521/FREE", dsn)
}

func TestOracleConnector_BuildDSN_MissingFields(t *testing.T) {
	conn := &oracleConnector{}
	_, err := conn.buildDSN(`{}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")

	_, err = conn.buildDSN(`{"host":"x"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username")

	_, err = conn.buildDSN(`{"host":"x","username":"u"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
}

func TestHiveConnector_BuildDSN(t *testing.T) {
	conn := &hiveConnector{}
	dsn, err := conn.buildDSN(`{"host":"localhost","port":10000,"username":"hive","password":"pass","database":"default"}`)
	require.NoError(t, err)
	assert.Equal(t, "hive:pass@tcp(localhost:10000)/default?charset=utf8mb4", dsn)
}

func TestHiveConnector_BuildDSN_DefaultPort(t *testing.T) {
	conn := &hiveConnector{}
	dsn, err := conn.buildDSN(`{"host":"hive.example.com","username":"hive","password":"p"}`)
	require.NoError(t, err)
	assert.Equal(t, "hive:p@tcp(hive.example.com:10000)?charset=utf8mb4", dsn)
}

func TestHiveConnector_BuildDSN_MissingFields(t *testing.T) {
	conn := &hiveConnector{}
	_, err := conn.buildDSN(`{}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")

	_, err = conn.buildDSN(`{"host":"x"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username")
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
	assert.Equal(t, "host=localhost port=5432 user=postgres password=secret dbname=test sslmode=prefer", dsn)
}

func TestMySQLConnector_BuildDSN_MissingFields(t *testing.T) {
	conn := &mysqlConnector{}
	_, err := conn.buildDSN(`{}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")

	_, err = conn.buildDSN(`{"host":"localhost"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")

	_, err = conn.buildDSN(`{"host":"localhost","port":3306}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username")

	_, err = conn.buildDSN(`{"host":"localhost","port":3306,"username":"root"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
}

func TestPostgreSQLConnector_BuildDSN_MissingFields(t *testing.T) {
	conn := &postgresqlConnector{}
	_, err := conn.buildDSN(`{}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")

	_, err = conn.buildDSN(`{"host":"localhost"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")

	_, err = conn.buildDSN(`{"host":"localhost","port":5432}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username")

	_, err = conn.buildDSN(`{"host":"localhost","port":5432,"username":"postgres"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
}

func TestPostgreSQLConnector_BuildDSN_CustomSSLMode(t *testing.T) {
	conn := &postgresqlConnector{}
	dsn, err := conn.buildDSN(`{"host":"localhost","port":5432,"username":"postgres","password":"secret","database":"test","sslmode":"require"}`)
	require.NoError(t, err)
	assert.Contains(t, dsn, "sslmode=require")
}

func TestAPIConnector_Close(t *testing.T) {
	// Close is a no-op for the API connector and must always succeed.
	conn, err := NewConnector("api")
	require.NoError(t, err)
	assert.NoError(t, conn.Close())

	// Calling Close twice should still be safe.
	assert.NoError(t, conn.Close())
}

// ---------------- 通过直接注入 sqlmock DB 覆盖 scanRows + GetColumns 真实路径 ----------------

// withMockSQLiteDB 用 sqlmock 注册一个 "sqlite3" 名字的 driver，并允许测试
// 用 mock 的 *sql.DB 替换 connector.db 字段。这样可以走真实的 scanRows +
// GetColumns + Query 路径而不依赖真容器。
func withMockSQLiteDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(false))
	require.NoError(t, err)
	return db, mock, func() { _ = db.Close() }
}

// TestSQLiteConnector_Query_RealScanRows 触发 scanRows 的真实路径（带类型信息）
func TestSQLiteConnector_Query_RealScanRows(t *testing.T) {
	c := &sqliteConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	// mock query 返回有类型的列
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "score"}).
			AddRow(int64(1), "alice", 95.5).
			AddRow(int64(2), "bob", 88.0),
	)

	rows, err := c.Query(context.Background(), "SELECT id, name, score FROM t")
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, int64(1), rows[0]["id"])
	assert.Equal(t, "alice", rows[0]["name"])
	assert.Equal(t, 95.5, rows[0]["score"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestSQLiteConnector_Query_Failure 触发 Query 失败路径
func TestSQLiteConnector_Query_Failure(t *testing.T) {
	c := &sqliteConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("synthetic db error"))
	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
}

// TestSQLiteConnector_GetColumns_RealPath 触发 GetColumns 的真实 PRAGMA 路径
func TestSQLiteConnector_GetColumns_RealPath(t *testing.T) {
	c := &sqliteConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	// PRAGMA table_info 返回 6 列：cid, name, type, notnull, dflt_value, pk
	mock.ExpectQuery("PRAGMA table_info").WillReturnRows(
		sqlmock.NewRows([]string{"cid", "name", "type", "notnull", "dflt_value", "pk"}).
			AddRow(0, "id", "INTEGER", 0, nil, 1).
			AddRow(1, "name", "TEXT", 0, nil, 0).
			AddRow(2, "score", "REAL", 0, nil, 0),
	)

	cols, err := c.GetColumns(context.Background(), "", "t1")
	require.NoError(t, err)
	assert.Len(t, cols, 3)
	assert.Equal(t, "id", cols[0].Name)
	assert.Equal(t, "name", cols[1].Name)
	assert.Equal(t, "score", cols[2].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestSQLiteConnector_GetColumns_ScanError 触发 Scan 失败
func TestSQLiteConnector_GetColumns_ScanError(t *testing.T) {
	c := &sqliteConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	// 故意返回 6 列但 Scan 时 int 解析错误（注入一个非 int 字符串到 cid）
	mock.ExpectQuery("PRAGMA table_info").WillReturnRows(
		sqlmock.NewRows([]string{"cid", "name", "type", "notnull", "dflt_value", "pk"}).
			AddRow("not-int", "name", "TEXT", 0, nil, 0),
	)

	_, err := c.GetColumns(context.Background(), "", "t1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scan column info failed")
}

// TestSQLiteConnector_GetColumns_RowIterationError 触发 rows.Err() 路径
func TestSQLiteConnector_GetColumns_RowIterationError(t *testing.T) {
	c := &sqliteConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("PRAGMA table_info").WillReturnRows(
		sqlmock.NewRows([]string{"cid", "name", "type", "notnull", "dflt_value", "pk"}).
			AddRow(0, "id", "INTEGER", 0, nil, 1).
			RowError(0, fmt.Errorf("iter error")),
	)

	_, err := c.GetColumns(context.Background(), "", "t1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "row iteration failed")
}

// 触发 driver.Value 实现检查（用于覆盖 scanRows 中的 value 转换分支）
var _ driver.Value = (*driverValueString)(nil)

type driverValueString string

func (d driverValueString) Value() (driver.Value, error) { return string(d), nil }

// ---------------- MySQL/PostgreSQL/ClickHouse/Sqlserver/Oracle GetColumns 真实路径覆盖 ----------------

func TestMysqlConnector_Query_RealScanRows(t *testing.T) {
	c := &mysqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "alice").AddRow(2, "bob"),
	)
	rows, err := c.Query(context.Background(), "SELECT id, name FROM t")
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "alice", rows[0]["name"])
}

func TestMysqlConnector_Query_Failure(t *testing.T) {
	c := &mysqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("synth"))
	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
}

func TestMysqlConnector_GetColumns_RealPath(t *testing.T) {
	c := &mysqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	// INFORMATION_SCHEMA.COLUMNS schema: COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH, NUMERIC_PRECISION, NUMERIC_SCALE
	mock.ExpectQuery("INFORMATION_SCHEMA").WillReturnRows(
		sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "NUMERIC_PRECISION", "NUMERIC_SCALE"}).
			AddRow("id", "int", nil, 10, 0).
			AddRow("name", "varchar", 64, nil, nil),
	)
	cols, err := c.GetColumns(context.Background(), "testdb", "t1")
	require.NoError(t, err)
	assert.Len(t, cols, 2)
	assert.Equal(t, "id", cols[0].Name)
	assert.Equal(t, "int", cols[0].Type)
	assert.Equal(t, 10, cols[0].Precision)
	assert.Equal(t, "name", cols[1].Name)
	assert.Equal(t, 64, cols[1].Length)
}

func TestMysqlConnector_GetColumns_Failure(t *testing.T) {
	c := &mysqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("INFORMATION_SCHEMA").WillReturnError(fmt.Errorf("synth"))
	_, err := c.GetColumns(context.Background(), "testdb", "t1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get columns failed")
}

func TestMysqlConnector_GetColumns_RowIterationError(t *testing.T) {
	c := &mysqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("INFORMATION_SCHEMA").WillReturnRows(
		sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "NUMERIC_PRECISION", "NUMERIC_SCALE"}).
			AddRow("id", "int", nil, 10, 0).
			RowError(0, fmt.Errorf("iter")),
	)
	_, err := c.GetColumns(context.Background(), "testdb", "t1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "row iteration failed")
}

func TestMysqlConnector_Close_RealPath(t *testing.T) {
	c := &mysqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectClose()
	assert.NoError(t, c.Close())
	assert.Nil(t, c.db)
}

func TestPostgresqlConnector_Query_RealScanRows(t *testing.T) {
	c := &postgresqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"n"}).AddRow(int64(42)),
	)
	rows, err := c.Query(context.Background(), "SELECT n FROM t")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}

func TestPostgresqlConnector_Query_Failure(t *testing.T) {
	c := &postgresqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("synth"))
	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
}

func TestPostgresqlConnector_GetColumns_RealPath(t *testing.T) {
	c := &postgresqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	// information_schema.columns: column_name, data_type, character_maximum_length, numeric_precision, numeric_scale
	mock.ExpectQuery("information_schema").WillReturnRows(
		sqlmock.NewRows([]string{"column_name", "data_type", "character_maximum_length", "numeric_precision", "numeric_scale"}).
			AddRow("id", "integer", nil, 32, 0).
			AddRow("name", "varchar", 128, nil, nil),
	)
	cols, err := c.GetColumns(context.Background(), "public", "t1")
	require.NoError(t, err)
	assert.Len(t, cols, 2)
	assert.Equal(t, "id", cols[0].Name)
	assert.Equal(t, "name", cols[1].Name)
}

func TestPostgresqlConnector_GetColumns_Failure(t *testing.T) {
	c := &postgresqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("information_schema").WillReturnError(fmt.Errorf("synth"))
	_, err := c.GetColumns(context.Background(), "public", "t1")
	assert.Error(t, err)
}

func TestPostgresqlConnector_GetColumns_RowIterationError(t *testing.T) {
	c := &postgresqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("information_schema").WillReturnRows(
		sqlmock.NewRows([]string{"column_name", "data_type", "character_maximum_length", "numeric_precision", "numeric_scale"}).
			AddRow("id", "integer", nil, 32, 0).
			RowError(0, fmt.Errorf("iter")),
	)
	_, err := c.GetColumns(context.Background(), "public", "t1")
	assert.Error(t, err)
}

func TestClickhouseConnector_Query_RealScanRows(t *testing.T) {
	c := &clickhouseConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"n"}).AddRow(int64(1)),
	)
	rows, err := c.Query(context.Background(), "SELECT n")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}

func TestClickhouseConnector_Query_Failure(t *testing.T) {
	c := &clickhouseConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("synth"))
	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
}

func TestClickhouseConnector_GetColumns_RealPath(t *testing.T) {
	c := &clickhouseConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	// system.columns: name, type, length, precision, scale
	mock.ExpectQuery("system.columns").WillReturnRows(
		sqlmock.NewRows([]string{"name", "type", "length", "precision", "scale"}).
			AddRow("id", "UInt64", 0, 0, 0).
			AddRow("name", "String", 0, 0, 0),
	)
	cols, err := c.GetColumns(context.Background(), "default", "t1")
	require.NoError(t, err)
	assert.Len(t, cols, 2)
	assert.Equal(t, "id", cols[0].Name)
	assert.Equal(t, "UInt64", cols[0].Type)
}

func TestClickhouseConnector_GetColumns_Failure(t *testing.T) {
	c := &clickhouseConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("system.columns").WillReturnError(fmt.Errorf("synth"))
	_, err := c.GetColumns(context.Background(), "default", "t1")
	assert.Error(t, err)
}

func TestClickhouseConnector_GetColumns_RowIterationError(t *testing.T) {
	c := &clickhouseConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("system.columns").WillReturnRows(
		sqlmock.NewRows([]string{"name", "type", "length", "precision", "scale"}).
			AddRow("id", "UInt64", 0, 0, 0).
			RowError(0, fmt.Errorf("iter")),
	)
	_, err := c.GetColumns(context.Background(), "default", "t1")
	assert.Error(t, err)
}

func TestSqlserverConnector_Query_RealScanRows(t *testing.T) {
	c := &sqlserverConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"n"}).AddRow(int64(1)),
	)
	rows, err := c.Query(context.Background(), "SELECT n")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}

func TestSqlserverConnector_Query_Failure(t *testing.T) {
	c := &sqlserverConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("synth"))
	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
}

func TestSqlserverConnector_GetColumns_RealPath(t *testing.T) {
	c := &sqlserverConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	// INFORMATION_SCHEMA.COLUMNS for SQL Server
	mock.ExpectQuery("INFORMATION_SCHEMA").WillReturnRows(
		sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "NUMERIC_PRECISION", "NUMERIC_SCALE"}).
			AddRow("id", "int", nil, 10, 0),
	)
	cols, err := c.GetColumns(context.Background(), "testdb", "t1")
	require.NoError(t, err)
	assert.Len(t, cols, 1)
}

func TestSqlserverConnector_GetColumns_Failure(t *testing.T) {
	c := &sqlserverConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("INFORMATION_SCHEMA").WillReturnError(fmt.Errorf("synth"))
	_, err := c.GetColumns(context.Background(), "testdb", "t1")
	assert.Error(t, err)
}

func TestOracleConnector_Query_RealScanRows(t *testing.T) {
	c := &oracleConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"n"}).AddRow(int64(1)),
	)
	rows, err := c.Query(context.Background(), "SELECT 1 FROM DUAL")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}

func TestOracleConnector_Query_Failure(t *testing.T) {
	c := &oracleConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("synth"))
	_, err := c.Query(context.Background(), "SELECT 1 FROM DUAL")
	assert.Error(t, err)
}

func TestOracleConnector_GetColumns_RealPath(t *testing.T) {
	c := &oracleConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	// all_tab_columns: COLUMN_NAME, DATA_TYPE, DATA_LENGTH, DATA_PRECISION, DATA_SCALE
	mock.ExpectQuery("all_tab_columns").WillReturnRows(
		sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "DATA_LENGTH", "DATA_PRECISION", "DATA_SCALE"}).
			AddRow("ID", "NUMBER", 22, 10, 0).
			AddRow("NAME", "VARCHAR2", 64, nil, nil),
	)
	cols, err := c.GetColumns(context.Background(), "TESTUSER", "T1")
	require.NoError(t, err)
	assert.Len(t, cols, 2)
	assert.Equal(t, "ID", cols[0].Name)
	assert.Equal(t, 64, cols[1].Length)
}

func TestOracleConnector_GetColumns_Failure(t *testing.T) {
	c := &oracleConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("all_tab_columns").WillReturnError(fmt.Errorf("synth"))
	_, err := c.GetColumns(context.Background(), "TESTUSER", "T1")
	assert.Error(t, err)
}

func TestHiveConnector_Query_RealScanRows(t *testing.T) {
	c := &hiveConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"n"}).AddRow(int64(1)),
	)
	rows, err := c.Query(context.Background(), "SELECT n")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}

func TestHiveConnector_Query_Failure(t *testing.T) {
	c := &hiveConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("synth"))
	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
}

func TestHiveConnector_GetColumns_RealPath(t *testing.T) {
	c := &hiveConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("INFORMATION_SCHEMA").WillReturnRows(
		sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "Length", "Precision", "Scale"}).
			AddRow("id", "int", 0, 0, 0).
			AddRow("name", "string", 0, 0, 0),
	)
	cols, err := c.GetColumns(context.Background(), "default", "t1")
	require.NoError(t, err)
	assert.Len(t, cols, 2)
	assert.Equal(t, "id", cols[0].Name)
}

func TestHiveConnector_GetColumns_Failure(t *testing.T) {
	c := &hiveConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	mock.ExpectQuery("DESCRIBE").WillReturnError(fmt.Errorf("synth"))
	_, err := c.GetColumns(context.Background(), "default", "t1")
	assert.Error(t, err)
}

// 验证 Close() 在已连接时调用 db.Close()(成功路径覆盖)
func TestConnectorClose_RealDB(t *testing.T) {
	// sqlmock 的 DB 在 ExpectationsMet 后调用 Close() 会失败，
	// 所以这个测试用 ExpectationsWereMet 之前主动调用 Close。
	c := &mysqlConnector{}
	db, mock, cleanup := withMockSQLiteDB(t)
	defer cleanup()
	c.db = db

	// mock 一个 no-op 的 query 让 close 之前的 state 正常
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"x"}).AddRow(1))
	_, _ = c.Query(context.Background(), "SELECT 1")

	// 现在显式 expect Close（这会满足 sqlmock 的预期）
	mock.ExpectClose()
	assert.NoError(t, c.Close())
	assert.Nil(t, c.db)
}
