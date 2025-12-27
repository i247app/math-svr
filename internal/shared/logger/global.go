package logger

import (
	"fmt"
	"log/slog"
	"os"
)

// Package-level logger functions for convenience
// These provide simple logging when you don't have a request-scoped logger

var defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

// Info logs an informational message using the default logger
func Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	defaultLogger.Info(msg)
}

// Error logs an error message using the default logger
func Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	defaultLogger.Error(msg)
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	defaultLogger.Debug(msg)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	defaultLogger.Warn(msg)
}
