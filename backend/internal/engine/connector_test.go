package engine

import (
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
