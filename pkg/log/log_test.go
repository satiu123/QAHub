package log_test

import (
	"context"
	"log/slog"
	"qahub/pkg/config"
	"qahub/pkg/log"
	"testing"

	"gopkg.in/natefinch/lumberjack.v2"
)

func TestInitLogger(t *testing.T) {
	cfg := &config.LogConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "multi",
		AddSource: true,
		InitialFields: map[string]any{
			"app": "test-app",
		},
		File: lumberjack.Logger{
			Filename:   "test.log",
			MaxSize:    5, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   // days
			Compress:   true, // disabled by default
		},
	}
	log.InitLogger(cfg)
	logger := slog.Default()
	logger.Info("This is a test log message", slog.String("key", "value"))
	logger.Debug("This is a debug log message", slog.Int("number", 42))
	logger.Warn("This is a warning log message")
	logger.Error("This is an error log message")
}

func TestLoggerWithContext(t *testing.T) {
	cfg := &config.LogConfig{
		Level:     "info",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}
	log.InitLogger(cfg)
	baseLogger := slog.Default()

	ctx := log.WithContext(context.TODO(), baseLogger)
	retrievedLogger := log.FromContext(ctx)

	retrievedLogger.Info("Log message from context logger", slog.String("context_key", "context_value"))
}
