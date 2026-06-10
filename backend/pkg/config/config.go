package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Logger   LoggerConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parse_time"`
	Loc             string        `mapstructure:"loc"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset, c.ParseTime, c.Loc)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type JWTConfig struct {
	Secret       string        `mapstructure:"secret"`
	ExpireHours  time.Duration `mapstructure:"expire_hours"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("COZYINSIGHT")
	// Allow Viper to coerce default-typed values (e.g. ${VAR:-3306} stays a string in YAML,
	// but the target struct field is int) by honoring mapstructure type tags.
	viper.SetTypeByDefaultValue(true)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Expand ${VAR:-default} placeholders that Viper does not natively resolve
	// when env vars are unset. Without this, port: ${VAR:-3306} stays as a literal
	// string and fails the int mapstructure decode.
	expanded, err := expandConfigPlaceholders(viper.AllSettings())
	if err != nil {
		return nil, fmt.Errorf("failed to expand config placeholders: %w", err)
	}

	var cfg Config
	if err := viper.MergeConfigMap(expanded); err != nil {
		return nil, fmt.Errorf("failed to merge expanded config: %w", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// expandConfigPlaceholders walks a Viper settings map and resolves ${VAR} / ${VAR:-default}
// patterns using os.Getenv. Strings that are pure numbers are left as strings so Viper can
// decode them into the target int field via SetTypeByDefaultValue.
func expandConfigPlaceholders(settings map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{}, len(settings))
	for k, v := range settings {
		switch val := v.(type) {
		case string:
			out[k] = os.Expand(val, func(name string) string {
				// ${NAME:-default}
				if idx := strings.Index(name, ":-"); idx >= 0 {
					key, def := name[:idx], name[idx+2:]
					if v, ok := os.LookupEnv("COZYINSIGHT_" + key); ok {
						return v
					}
					return def
				}
				if v, ok := os.LookupEnv("COZYINSIGHT_" + name); ok {
					return v
				}
				return ""
			})
		case map[string]interface{}:
			nested, err := expandConfigPlaceholders(val)
			if err != nil {
				return nil, err
			}
			out[k] = nested
		default:
			out[k] = v
		}
	}
	return out, nil
}
