package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cozy-insight/pkg/config"
)

func TestNew_Default(t *testing.T) {
	cfg := config.LoggerConfig{
		Level: "info",
	}
	log := New(cfg)
	require.NotNil(t, log)
	assert.NotNil(t, log.Core())
	log.Sync() //nolint:errcheck
}

func TestNew_WithFile(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:      "debug",
		Filename:   "",
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 3,
	}
	log := New(cfg)
	require.NotNil(t, log)
	log.Sync() //nolint:errcheck
}
