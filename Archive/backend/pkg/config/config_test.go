package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func() string
		wantErr     bool
		checkConfig func(*testing.T, *Config)
	}{
		{
			name: "Valid config file",
			setupFile: func() string {
				content := `
server:
  port: 8080
  mode: "release"

database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  database: "testdb"
  charset: "utf8mb4"
  parse_time: true
  loc: "Local"

calcite:
  avatica_url: "http://localhost:8765"
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 3600000000000

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

logger:
  level: "info"
  filename: "logs/app.log"
  max_size: 100
  max_age: 7
  max_backups: 3

jwt:
  secret: "test-secret-key"
`
				tmpFile, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)
				_, err = tmpFile.WriteString(content)
				require.NoError(t, err)
				tmpFile.Close()
				return tmpFile.Name()
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 8080, cfg.Server.Port)
				assert.Equal(t, "release", cfg.Server.Mode)
				assert.Equal(t, "mysql", cfg.Database.Driver)
				assert.Equal(t, "localhost", cfg.Database.Host)
				assert.Equal(t, 3306, cfg.Database.Port)
				assert.Equal(t, "test-secret-key", cfg.JWT.Secret)
				assert.Equal(t, "info", cfg.Logger.Level)
			},
		},
		{
			name: "File does not exist",
			setupFile: func() string {
				return "/nonexistent/path/config.yaml"
			},
			wantErr: true,
		},
		{
			name: "Invalid YAML syntax",
			setupFile: func() string {
				content := `
server:
  port: 8080
  invalid yaml syntax [[[
`
				tmpFile, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)
				_, err = tmpFile.WriteString(content)
				require.NoError(t, err)
				tmpFile.Close()
				return tmpFile.Name()
			},
			wantErr: true,
		},
		{
			name: "Empty config file",
			setupFile: func() string {
				tmpFile, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)
				tmpFile.Close()
				return tmpFile.Name()
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				// Should have zero values
				assert.Equal(t, 0, cfg.Server.Port)
				assert.Equal(t, "", cfg.Server.Mode)
			},
		},
		{
			name: "Partial config",
			setupFile: func() string {
				content := `
server:
  port: 9090

jwt:
  secret: "my-secret"
`
				tmpFile, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)
				_, err = tmpFile.WriteString(content)
				require.NoError(t, err)
				tmpFile.Close()
				return tmpFile.Name()
			},
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 9090, cfg.Server.Port)
				assert.Equal(t, "my-secret", cfg.JWT.Secret)
				// Other fields should have zero values
				assert.Equal(t, "", cfg.Database.Host)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupFile()

			// Cleanup temporary file after test
			if filepath.IsAbs(configPath) && filepath.Dir(configPath) == os.TempDir() {
				defer os.Remove(configPath)
			}

			cfg, err := LoadConfig(configPath)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				if tt.checkConfig != nil {
					tt.checkConfig(t, cfg)
				}
			}
		})
	}
}

func TestGlobalConfig(t *testing.T) {
	content := `
server:
  port: 7777
  mode: "debug"

jwt:
  secret: "global-secret"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check GlobalConfig is set
	assert.NotNil(t, GlobalConfig)
	assert.Equal(t, cfg, GlobalConfig)
	assert.Equal(t, 7777, GlobalConfig.Server.Port)
	assert.Equal(t, "global-secret", GlobalConfig.JWT.Secret)
}

func TestConfigStructs(t *testing.T) {
	t.Run("ServerConfig", func(t *testing.T) {
		sc := ServerConfig{
			Port: 8080,
			Mode: "release",
		}
		assert.Equal(t, 8080, sc.Port)
		assert.Equal(t, "release", sc.Mode)
	})

	t.Run("DatabaseConfig", func(t *testing.T) {
		dc := DatabaseConfig{
			Driver:    "mysql",
			Host:      "localhost",
			Port:      3306,
			Username:  "user",
			Password:  "pass",
			Database:  "db",
			Charset:   "utf8mb4",
			ParseTime: true,
			Loc:       "Local",
		}
		assert.Equal(t, "mysql", dc.Driver)
		assert.Equal(t, 3306, dc.Port)
		assert.True(t, dc.ParseTime)
	})

	t.Run("RedisConfig", func(t *testing.T) {
		rc := RedisConfig{
			Host:     "redis-server",
			Port:     6379,
			Password: "secret",
			DB:       1,
		}
		assert.Equal(t, "redis-server", rc.Host)
		assert.Equal(t, 6379, rc.Port)
		assert.Equal(t, 1, rc.DB)
	})
}
