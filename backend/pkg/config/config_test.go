package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	content := `
server:
  port: 9999
  mode: test
database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  username: root
  password: secret
  database: testdb
  charset: utf8mb4
  parse_time: true
  loc: Local
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 3600000000000
redis:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0
logger:
  level: debug
  filename: ""
  max_size: 100
  max_age: 7
  max_backups: 3
jwt:
  secret: test-secret
  expire_hours: 2
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, "test", cfg.Server.Mode)
	assert.Equal(t, "mysql", cfg.Database.Driver)
	assert.Equal(t, "testdb", cfg.Database.Database)
	assert.Equal(t, "test-secret", cfg.JWT.Secret)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Username:   "root",
		Password:   "secret",
		Host:       "127.0.0.1",
		Port:       3306,
		Database:   "testdb",
		Charset:    "utf8mb4",
		ParseTime:  true,
		Loc:        "Local",
	}
	expected := "root:secret@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=true&loc=Local"
	assert.Equal(t, expected, cfg.DSN())
}
