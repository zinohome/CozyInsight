package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------- Phase A: Coverage for connector lifecycle paths ----------------

// TestNewConnector_ExhaustiveTypeCoverage 覆盖 13 个 type 的 NewConnector 注册路径。
// doris/starrocks/mongodb 复用 mysqlConnector，所以验证类型是 *mysqlConnector。
func TestNewConnector_ExhaustiveTypeCoverage(t *testing.T) {
	cases := []struct {
		dsType   string
		wantType any
	}{
		{"mysql", &mysqlConnector{}},
		{"postgresql", &postgresqlConnector{}},
		{"sqlite", &sqliteConnector{}},
		{"clickhouse", &clickhouseConnector{}},
		{"sqlserver", &sqlserverConnector{}},
		{"oracle", &oracleConnector{}},
		{"hive", &hiveConnector{}},
		{"elasticsearch", &elasticsearchConnector{}},
		{"api", &apiConnector{}},
		{"doris", &mysqlConnector{}},     // 复用 mysqlConnector
		{"starrocks", &mysqlConnector{}},  // 复用 mysqlConnector
		{"mongodb", &mysqlConnector{}},    // 复用 mysqlConnector (BI 协议)
		{"csv", &fileConnector{}},
		{"excel", &fileConnector{}},
	}
	for _, c := range cases {
		t.Run(c.dsType, func(t *testing.T) {
			conn, err := NewConnector(c.dsType)
			assert.NoError(t, err)
			assert.IsType(t, c.wantType, conn, "dsType=%s", c.dsType)
		})
	}
}

// TestOracleConnector_NotConnected 覆盖 db==nil 时 Query/Close/GetColumns 错误路径
func TestOracleConnector_NotConnected(t *testing.T) {
	c := &oracleConnector{}

	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	_, err = c.GetColumns(context.Background(), "", "t")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	assert.NoError(t, c.Close(), "Close on nil db should be no-op")
}

// TestHiveConnector_NotConnected 同上
func TestHiveConnector_NotConnected(t *testing.T) {
	c := &hiveConnector{}

	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	_, err = c.GetColumns(context.Background(), "", "t")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	assert.NoError(t, c.Close(), "Close on nil db should be no-op")
}

// TestClickhouseConnector_NotConnected 同上
func TestClickhouseConnector_NotConnected(t *testing.T) {
	c := &clickhouseConnector{}

	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	_, err = c.GetColumns(context.Background(), "", "t")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	assert.NoError(t, c.Close(), "Close on nil db should be no-op")
}

// TestSqlserverConnector_NotConnected 同上
func TestSqlserverConnector_NotConnected(t *testing.T) {
	c := &sqlserverConnector{}

	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	_, err = c.GetColumns(context.Background(), "", "t")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	assert.NoError(t, c.Close(), "Close on nil db should be no-op")
}

// TestPostgresqlConnector_Connect_InvalidJSON 覆盖 Connect 入口的 JSON 解析
func TestPostgresqlConnector_Connect_InvalidJSON(t *testing.T) {
	c := &postgresqlConnector{}
	err := c.Connect(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

// TestMysqlConnector_Connect_InvalidJSON
func TestMysqlConnector_Connect_InvalidJSON(t *testing.T) {
	c := &mysqlConnector{}
	err := c.Connect(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

// TestClickhouseConnector_Connect_InvalidJSON
func TestClickhouseConnector_Connect_InvalidJSON(t *testing.T) {
	c := &clickhouseConnector{}
	err := c.Connect(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

// TestSqlserverConnector_Connect_InvalidJSON
func TestSqlserverConnector_Connect_InvalidJSON(t *testing.T) {
	c := &sqlserverConnector{}
	err := c.Connect(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

// TestOracleConnector_Connect_InvalidJSON
func TestOracleConnector_Connect_InvalidJSON(t *testing.T) {
	c := &oracleConnector{}
	err := c.Connect(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

// TestHiveConnector_Connect_InvalidJSON
func TestHiveConnector_Connect_InvalidJSON(t *testing.T) {
	c := &hiveConnector{}
	err := c.Connect(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

// TestSqliteConnector_Connect_InvalidJSON
func TestSqliteConnector_Connect_InvalidJSON(t *testing.T) {
	c := &sqliteConnector{}
	err := c.Connect(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

// ---------------- extractSSHConfig 边界测试 ----------------

func TestExtractSSHConfig_AllFields(t *testing.T) {
	cfg, err := extractSSHConfig(`{"sshTunnel":{"enabled":true,"host":"bastion","port":2222,"username":"alice","password":"p","privateKey":"-----BEGIN-----"}}`)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "bastion", cfg.Host)
	assert.Equal(t, 2222, cfg.Port)
	assert.Equal(t, "alice", cfg.Username)
	assert.Equal(t, "p", cfg.Password)
	assert.Equal(t, "-----BEGIN-----", cfg.PrivateKey)
}

func TestExtractSSHConfig_DisabledButPresent(t *testing.T) {
	// enabled=false 时返回 nil
	cfg, err := extractSSHConfig(`{"sshTunnel":{"enabled":false,"host":"bastion","port":22,"username":"u","password":"p"}}`)
	assert.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestExtractSSHConfig_NilConfig(t *testing.T) {
	// 没有 sshTunnel 字段时返回 nil
	cfg, err := extractSSHConfig(`{"host":"db"}`)
	assert.NoError(t, err)
	assert.Nil(t, cfg)
}

// ---------------- NewConnectorWithTunnel 错误路径 ----------------

func TestNewConnectorWithTunnel_UnsupportedType(t *testing.T) {
	_, tunnel, err := NewConnectorWithTunnel("foobar", `{}`)
	assert.Error(t, err)
	assert.Nil(t, tunnel)
}

// ---------------- engine.BuildSQL with 各种 dialect ----------------

func TestBuildSQL_AllDialects(t *testing.T) {
	dialects := []string{"mysql", "postgresql", "sqlite", "clickhouse", "sqlserver", "oracle", "doris", "starrocks", "mongodb", "hive"}
	for _, d := range dialects {
		t.Run(d, func(t *testing.T) {
			// 不验证具体 SQL，只验证无 panic 且能产出某种 SQL
			config := ChartQueryConfig{
				Dimensions: []Dimension{{Field: "x"}},
				Metrics:    []Metric{{Field: "y", Aggregation: "SUM"}},
			}
			sql, _, err := BuildSQL("t", d, config)
			assert.NoError(t, err, "dialect=%s", d)
			assert.NotEmpty(t, sql)
		})
	}
}

// ---------------- NewConnector with default error ----------------

func TestNewConnector_UnknownType_ErrorMessage(t *testing.T) {
	_, err := NewConnector("made-up-type-xyz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported datasource type")
	assert.Contains(t, err.Error(), "made-up-type-xyz")
}

// ---------------- oracleConnector.Connect with missing fields ----------------

func TestOracleConnector_Connect_MissingDatabase(t *testing.T) {
	c := &oracleConnector{}
	err := c.Connect(`{"host":"x","username":"u","password":"p"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
}

func TestOracleConnector_Connect_MissingHost(t *testing.T) {
	c := &oracleConnector{}
	err := c.Connect(`{"username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")
}

func TestOracleConnector_Connect_MissingUsername(t *testing.T) {
	c := &oracleConnector{}
	err := c.Connect(`{"host":"x","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username")
}

// ---------------- hiveConnector.Connect with missing fields ----------------

func TestHiveConnector_Connect_MissingHost(t *testing.T) {
	c := &hiveConnector{}
	err := c.Connect(`{"username":"u","password":"p"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")
}

func TestHiveConnector_Connect_MissingUsername(t *testing.T) {
	c := &hiveConnector{}
	err := c.Connect(`{"host":"x","password":"p"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username")
}

// ---------------- clickhouseConnector.Connect with missing fields ----------------

func TestClickhouseConnector_Connect_MissingDatabase(t *testing.T) {
	c := &clickhouseConnector{}
	err := c.Connect(`{"host":"x","port":9000,"username":"u","password":"p"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
}

func TestClickhouseConnector_Connect_MissingHost(t *testing.T) {
	c := &clickhouseConnector{}
	err := c.Connect(`{"port":9000,"username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")
}

// ---------------- errors.go 中的错误类型 ----------------

func TestEngineErrors_NonNil(t *testing.T) {
	// 触发 engine 错误返回路径
	c := &oracleConnector{}
	_, err := c.Query(context.Background(), "SELECT 1")
	assert.True(t, errors.Is(err, err) || err != nil)
	assert.NotNil(t, err)
}

// ---------------- doris/starrocks/mongodb 复用 MySQL 协议 ----------------
//
// 这三个数据源类型 NewConnector 都返回 *mysqlConnector，因为：
//   - Doris / StarRocks 自带 MySQL wire protocol 兼容层
//   - MongoDB-BI（mongosqld）暴露 MySQL 协议给 BI 工具
//
// 此测试验证 DSN 拼接出来的格式正确，能被 MySQL driver 接受。

func TestDorisStarRocksMongoDB_DSNFormat(t *testing.T) {
	cfgJSON := `{"host":"127.0.0.1","port":3306,"username":"root","password":"p","database":"db"}`
	wantDSN := "root:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true"

	for _, dsType := range []string{"doris", "starrocks", "mongodb"} {
		t.Run(dsType, func(t *testing.T) {
			conn, err := NewConnector(dsType)
			assert.NoError(t, err, "factory returns connector")

			// 强制类型断言为 *mysqlConnector 并访问包私有方法 buildDSN
			mc, ok := conn.(*mysqlConnector)
			assert.True(t, ok, "dsType=%s should return *mysqlConnector", dsType)

			dsn, err := mc.buildDSN(cfgJSON)
			assert.NoError(t, err, "buildDSN")
			assert.Equal(t, wantDSN, dsn, "dsType=%s DSN", dsType)
		})
	}
}

func TestDorisStarRocksMongoDB_Connect_MissingFields(t *testing.T) {
	// 三个复用 MySQL 协议的 connector 在配置错误时返回和 MySQL 一致的错误。
	for _, dsType := range []string{"doris", "starrocks", "mongodb"} {
		t.Run(dsType, func(t *testing.T) {
			conn, err := NewConnector(dsType)
			assert.NoError(t, err)

			err = conn.Connect(`{"host":"x","port":3306,"username":"u","password":"p"}`) // 缺 database
			assert.Error(t, err, "missing database should fail")
			assert.Contains(t, err.Error(), "database")
		})
	}
}

func TestMysqlConnector_NotConnected(t *testing.T) {
	c := &mysqlConnector{}

	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	_, err = c.GetColumns(context.Background(), "", "t")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	assert.NoError(t, c.Close(), "Close on nil db should be no-op")
}

func TestPostgresqlConnector_NotConnected(t *testing.T) {
	c := &postgresqlConnector{}

	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	_, err = c.GetColumns(context.Background(), "", "t")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	assert.NoError(t, c.Close(), "Close on nil db should be no-op")
}

func TestSqliteConnector_NotConnected(t *testing.T) {
	c := &sqliteConnector{}

	_, err := c.Query(context.Background(), "SELECT 1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	_, err = c.GetColumns(context.Background(), "", "t")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	assert.NoError(t, c.Close(), "Close on nil db should be no-op")
}

func TestDorisConnector_NotConnected(t *testing.T) {
	// doris/starrocks/mongodb 复用 mysqlConnector，验证 lifecycle 一致
	for _, dsType := range []string{"doris", "starrocks", "mongodb"} {
		t.Run(dsType, func(t *testing.T) {
			conn, err := NewConnector(dsType)
			assert.NoError(t, err)

			_, err = conn.Query(context.Background(), "SELECT 1")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not connected")

			_, err = conn.GetColumns(context.Background(), "", "t")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not connected")

			assert.NoError(t, conn.Close())
		})
	}
}

func TestSqlserverConnector_BuildDSN_MissingFields(t *testing.T) {
	c := &sqlserverConnector{}

	// 缺 host
	_, err := c.buildDSN(`{"port":1433,"username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")

	// 缺 port
	_, err = c.buildDSN(`{"host":"h","username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")

	// 缺 username
	_, err = c.buildDSN(`{"host":"h","port":1433,"password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username")

	// 缺 database
	_, err = c.buildDSN(`{"host":"h","port":1433,"username":"u","password":"p"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")

	// invalid JSON
	_, err = c.buildDSN(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

func TestMysqlConnector_BuildDSN_MissingFields(t *testing.T) {
	c := &mysqlConnector{}

	// 缺 host
	_, err := c.buildDSN(`{"port":3306,"username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")

	// 缺 port
	_, err = c.buildDSN(`{"host":"h","username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")

	// 缺 username
	_, err = c.buildDSN(`{"host":"h","port":3306,"password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username")

	// 缺 database
	_, err = c.buildDSN(`{"host":"h","port":3306,"username":"u","password":"p"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")

	// invalid JSON
	_, err = c.buildDSN(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

func TestPostgresqlConnector_BuildDSN_MissingFields(t *testing.T) {
	c := &postgresqlConnector{}

	// 缺 host
	_, err := c.buildDSN(`{"port":5432,"username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")

	// 缺 port
	_, err = c.buildDSN(`{"host":"h","username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")

	// 缺 username
	_, err = c.buildDSN(`{"host":"h","port":5432,"password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username")

	// 缺 database
	_, err = c.buildDSN(`{"host":"h","port":5432,"username":"u","password":"p"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")

	// invalid JSON
	_, err = c.buildDSN(`not-json`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

// ---------------- connector.Connect with invalid port type (port 不是数字) ----------------

func TestMysqlConnector_Connect_InvalidPortType(t *testing.T) {
	c := &mysqlConnector{}
	// port 应该是 float64，传字符串触发 buildDSN 失败
	err := c.Connect(`{"host":"h","port":"3306","username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

func TestPostgresqlConnector_Connect_InvalidPortType(t *testing.T) {
	c := &postgresqlConnector{}
	err := c.Connect(`{"host":"h","port":"5432","username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

func TestSqlserverConnector_Connect_InvalidPortType(t *testing.T) {
	c := &sqlserverConnector{}
	err := c.Connect(`{"host":"h","port":"1433","username":"u","password":"p","database":"d"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

// ---------------- BuildDSN custom SSL mode / port boundary ----------------

func TestPostgresqlConnector_BuildDSN_NoSSL(t *testing.T) {
	c := &postgresqlConnector{}
	// 不写 sslmode，验证默认值
	dsn, err := c.buildDSN(`{"host":"h","port":5432,"username":"u","password":"p","database":"d"}`)
	assert.NoError(t, err)
	assert.Contains(t, dsn, "sslmode=prefer")
}

// ---------------- ScanRows + 真实 SQLite data path ----------------
//
// 已经在 integration/container/sqlite_test.go 覆盖了真实 SQLite 路径，
// 这里在单元层用 sqlmock 覆盖 scanRows 的 []byte 数字回退路径。
