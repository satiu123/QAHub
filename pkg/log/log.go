package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"qahub/pkg/config"
)

type contextKey string

const loggerKey = contextKey("logger")

// InitLogger 初始化全局 logger
func InitLogger(cfg *config.LogConfig) {
	handlerOpts := &slog.HandlerOptions{
		Level:     levelFromString(cfg.Level),
		AddSource: cfg.AddSource,
	}
	writer := writerFromString(cfg.Output, cfg)
	handler := formatFromString(writer, cfg.Format, handlerOpts)
	var logger *slog.Logger
	if len(cfg.InitialFields) > 0 {
		initialAttrs := make([]any, 0, len(cfg.InitialFields))
		for k, v := range cfg.InitialFields {
			initialAttrs = append(initialAttrs, slog.Any(k, v))
		}
		logger = slog.New(handler).With(slog.Group("service_context", initialAttrs...))
	} else {
		logger = slog.New(handler)
	}
	slog.SetDefault(logger)
}

// WithContext 将 logger 存入 context
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext 从 context 中获取 logger，如果不存在则返回默认 logger
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func levelFromString(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
func writerFromString(outputStr string, fileCfg *config.LogConfig) *io.Writer {
	var w io.Writer
	switch outputStr {
	case "stdout":
		w = io.Writer(os.Stdout)
	case "stderr":
		w = io.Writer(os.Stderr)
	case "file":
		w = &fileCfg.File
	case "multi":
		w = io.MultiWriter(os.Stdout, &fileCfg.File)
	default:
		w = io.Writer(os.Stdout)
	}
	return &w
}

func formatFromString(w *io.Writer, formatStr string, handlerOpts *slog.HandlerOptions) slog.Handler {
	switch formatStr {
	case "json":
		return slog.NewJSONHandler(*w, handlerOpts)
	default:
		return slog.NewTextHandler(*w, handlerOpts)
	}
}
