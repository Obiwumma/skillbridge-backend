// Package logger provides structured JSON logging for observability.
package logger

import (
	"context"
	"os"
	"log/slog"
)

// Log is the globally shared structured logger instance.
var Log *slog.Logger

// Init configures the global logger to print JSON formatted logs to standard output.
func Init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Log = slog.New(handler)
	slog.SetDefault(Log)
}

// Info records structured logs at info severity.
func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

// Error records structured logs at error severity, ensuring the error is mapped first.
func Error(msg string, err error, args ...any) {
	newArgs := append([]any{slog.Any("error", err)}, args...)
	Log.Error(msg, newArgs...)
}

// Warn records structured logs at warning severity.
func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

// Debug records structured logs at debug severity.
func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

// WithContext returns a logger instance bound to the request context.
func WithContext(ctx context.Context) *slog.Logger {
	return Log
}
