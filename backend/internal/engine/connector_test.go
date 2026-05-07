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
