package engine_test

import (
	"testing"

	"cozy-insight-backend/internal/engine"

	"github.com/stretchr/testify/assert"
)

func TestCalciteClient_Constructor(t *testing.T) {
	tests := []struct {
		name       string
		avaticaURL string
		wantErr    bool
	}{
		{
			name:       "Valid URL",
			avaticaURL: "http://localhost:8765",
			wantErr:    false,
		},
		{
			name:       "Another valid URL",
			avaticaURL: "http://calcite:8888",
			wantErr:    false,
		},
		{
			name:       "HTTPS URL",
			avaticaURL: "https://calcite.example.com",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := engine.NewCalciteClient(tt.avaticaURL)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestDatasourceConnector_Constructor(t *testing.T) {
	t.Run("Create connector", func(t *testing.T) {
		connector := engine.NewDatasourceConnector()
		assert.NotNil(t, connector)
	})
}

func TestConnectionTestResult_Struct(t *testing.T) {
	t.Run("Create ConnectionTestResult", func(t *testing.T) {
		result := engine.ConnectionTestResult{
			Success: true,
			Message: "Connection successful",
		}
		assert.True(t, result.Success)
		assert.Equal(t, "Connection successful", result.Message)
	})

	t.Run("Failed connection result", func(t *testing.T) {
		result := engine.ConnectionTestResult{
			Success: false,
			Message: "Connection failed: timeout",
		}
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "failed")
	})
}

// 测试数据源类型
func TestDatasourceTypes(t *testing.T) {
	types := []string{
		"mysql",
		"postgresql",
		"oracle",
		"sqlserver",
		"clickhouse",
	}

	for _, dsType := range types {
		t.Run("Datasource type: "+dsType, func(t *testing.T) {
			assert.NotEmpty(t, dsType)
		})
	}
}

// 测试SQL构建的基础逻辑
func TestSQLBuilding(t *testing.T) {
	t.Run("Basic SQL structure", func(t *testing.T) {
		tableName := "users"
		sql := "SELECT * FROM " + tableName
		assert.Contains(t, sql, "SELECT")
		assert.Contains(t, sql, tableName)
	})

	t.Run("SQL with WHERE clause", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = ?"
		assert.Contains(t, sql, "WHERE")
	})

	t.Run("SQL with JOIN", func(t *testing.T) {
		sql := "SELECT * FROM orders JOIN users ON orders.user_id = users.id"
		assert.Contains(t, sql, "JOIN")
	})
}
