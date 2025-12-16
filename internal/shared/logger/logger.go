package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"
)

// Context keys for storing logger and session info in context
// ------------------------------------------------------------

type loggerKeyType string
type tokenKeyType string
type useridKeyType string

const (
	loggerKey = loggerKeyType("logger")
	tokenKey  = tokenKeyType("token")
	useridKey = useridKeyType("userid")
)

// Context helper functions
// ------------------------------------------------------------

func WithLogger(ctx context.Context, logger *logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func GetLogger(ctx context.Context) *logger {
	val := ctx.Value(loggerKey)
	if val == nil {
		return nil
	}
	return val.(*logger)
}

type logger struct {
	slogger *slog.Logger
	request *http.Request
	outFile *os.File
	ctx     context.Context
}

// NewRequestScopedLogger creates a new request-scoped logger instance
func NewRequestScopedLogger(r *http.Request, outFilePath string) *logger {
	var outFile *os.File
	var writer io.Writer

	// If outFilePath is provided, open the file for writing
	if outFilePath != "" {
		f, err := os.OpenFile(outFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// If file opening fails, fallback to stdout
			writer = os.Stdout
		} else {
			outFile = f
			// Write to both file and stdout
			writer = io.MultiWriter(f, os.Stdout)
		}
	} else {
		// Default to stdout
		writer = os.Stdout
	}

	// Create custom handler
	handler := newCustomHandler(writer)

	// Create slog logger with custom handler
	slogger := slog.New(handler)

	// Extract session info from request
	token, userid := extractSessionInfoFromRequest(r)

	// Create context with session info
	ctx := withSessionInfo(r.Context(), token, userid)

	return &logger{
		slogger: slogger,
		request: r,
		outFile: outFile,
		ctx:     ctx,
	}
}

// Close closes the logger's file handle if it exists
func (l *logger) Close() error {
	if l.outFile != nil {
		return l.outFile.Close()
	}
	return nil
}

// Info logs an informational message
func (l *logger) Info(args ...any) {
	msg := fmt.Sprint(args...)
	l.log(l.ctx, slog.LevelInfo, msg)
}

// Infof logs a formatted informational message
func (l *logger) Infof(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.log(l.ctx, slog.LevelInfo, msg)
}

// Error logs an error message
func (l *logger) Error(args ...any) {
	msg := fmt.Sprint(args...)
	l.log(l.ctx, slog.LevelError, msg)
}

// Errorf logs a formatted error message
func (l *logger) Errorf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.log(l.ctx, slog.LevelError, msg)
}

// Debug logs a debug message
func (l *logger) Debug(args ...any) {
	msg := fmt.Sprint(args...)
	l.log(l.ctx, slog.LevelDebug, msg)
}

// Debugf logs a formatted debug message
func (l *logger) Debugf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.log(l.ctx, slog.LevelDebug, msg)
}

// Warn logs a warning message
func (l *logger) Warn(args ...any) {
	msg := fmt.Sprint(args...)
	l.log(l.ctx, slog.LevelWarn, msg)
}

// Warnf logs a formatted warning message
func (l *logger) Warnf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.log(l.ctx, slog.LevelWarn, msg)
}

// log is a helper method that creates a log record with the correct caller information
func (l *logger) log(ctx context.Context, level slog.Level, msg string) {
	// Get the caller's program counter (PC)
	// Skip 2 frames: this function and the public logging method (Info, Error, etc.)
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])

	// Create a new record with the correct PC
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])

	// Call the handler directly
	_ = l.slogger.Handler().Handle(ctx, r)
}
