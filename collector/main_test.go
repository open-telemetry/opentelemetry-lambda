package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestInitLogger(t *testing.T) {
	logger := initLogger()
	assert.Equal(t, zapcore.WarnLevel, logger.Level())
}

func TestInitLogger_WithEnv(t *testing.T) {
	os.Setenv("OPENTELEMETRY_EXTENSION_LOG_LEVEL", "debug")
	logger := initLogger()
	assert.Equal(t, zapcore.DebugLevel, logger.Level())
	os.Unsetenv("OPENTELEMETRY_EXTENSION_LOG_LEVEL")
}
