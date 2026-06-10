package engine

import (
	"context"
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
