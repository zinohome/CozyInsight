package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel zapcore.Level
	}{
		{
			name:          "Debug level",
			level:         "debug",
			expectedLevel: zapcore.DebugLevel,
		},
		{
			name:          "Info level",
			level:         "info",
			expectedLevel: zapcore.InfoLevel,
		},
		{
			name:          "Warn level",
			level:         "warn",
			expectedLevel: zapcore.WarnLevel,
		},
		{
			name:          "Error level",
			level:         "error",
			expectedLevel: zapcore.ErrorLevel,
		},
		{
			name:          "Default level (unknown)",
			level:         "unknown",
			expectedLevel: zapcore.InfoLevel, // Default
		},
		{
			name:          "Empty level",
			level:         "",
			expectedLevel: zapcore.InfoLevel, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize logger
			InitLogger(tt.level)

			// Verify logger is not nil
			assert.NotNil(t, Log)

			// Verify global logger is set
			globalLogger := zap.L()
			assert.NotNil(t, globalLogger)

			// Verify logger can be used
			Log.Info("test log message")
			Log.Debug("debug message")
			Log.Warn("warn message")
			Log.Error("error message")

			// Basic verification that logger was created
			assert.NotNil(t, Log)
		})
	}
}

func TestLoggerFunctionality(t *testing.T) {
	// Initialize logger
	InitLogger("debug")

	t.Run("LoggerNotNil", func(t *testing.T) {
		assert.NotNil(t, Log)
	})

	t.Run("CanLogAtDifferentLevels", func(t *testing.T) {
		// These should not panic
		assert.NotPanics(t, func() {
			Log.Debug("debug message")
		})

		assert.NotPanics(t, func() {
			Log.Info("info message")
		})

		assert.NotPanics(t, func() {
			Log.Warn("warn message")
		})

		assert.NotPanics(t, func() {
			Log.Error("error message")
		})
	})

	t.Run("CanLogWithFields", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Log.Info("message with fields",
				zap.String("key", "value"),
				zap.Int("count", 42),
			)
		})
	})
}

func TestLoggerReinitialization(t *testing.T) {
	// Initialize with debug level
	InitLogger("debug")
	firstLogger := Log

	// Reinitialize with error level
	InitLogger("error")
	secondLogger := Log

	// Loggers should be different instances
	assert.NotEqual(t, firstLogger, secondLogger)
	assert.NotNil(t, Log)
}
