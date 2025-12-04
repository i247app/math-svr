package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"math-ai.com/math-ai/internal/session"
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

// withSessionInfo adds session info (token, userid) to context for logging
func withSessionInfo(ctx context.Context, token, userid string) context.Context {
	ctx = context.WithValue(ctx, tokenKey, token)
	ctx = context.WithValue(ctx, useridKey, userid)
	return ctx
}

// extractSessionInfo retrieves token and userid from context
func extractSessionInfo(ctx context.Context) (token string, userid string) {
	token = "anon"
	userid = "anon"

	if t, ok := ctx.Value(tokenKey).(string); ok && t != "" {
		token = t
	}
	if u, ok := ctx.Value(useridKey).(string); ok && u != "" {
		userid = u
	}

	return
}

// Custom slog.Handler implementation
// ------------------------------------------------------------

// customHandler implements slog.Handler with our custom format
type customHandler struct {
	writer io.Writer
	attrs  []slog.Attr
	groups []string
}

// newCustomHandler creates a new custom handler that writes to the given writer
func newCustomHandler(w io.Writer) *customHandler {
	return &customHandler{
		writer: w,
		attrs:  make([]slog.Attr, 0),
		groups: make([]string, 0),
	}
}

// Enabled reports whether the handler handles records at the given level
func (h *customHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Enable all levels
	return true
}

// Handle formats and writes the log record
func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract session info from context
	token, userid := extractSessionInfo(ctx)

	// Get caller information (filename and line)
	// We need to skip frames to get the actual caller
	var file string
	var line int

	// Try to get source from record first
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		file = filepath.Base(f.File)
		line = f.Line
	} else {
		// Fallback: walk the stack to find the caller
		var pcs [1]uintptr
		runtime.Callers(5, pcs[:])
		fs := runtime.CallersFrames(pcs[:])
		f, _ := fs.Next()
		file = filepath.Base(f.File)
		line = f.Line
	}

	// Format timestamp with microseconds: 2025/12/04 04:18:38.151018
	timestamp := r.Time.Format("2006/01/02 15:04:05.000000")

	// Format level
	level := r.Level.String()

	// Build the log message
	logMsg := fmt.Sprintf("%s %s:%d [%s] [%s] %s: %s\n",
		timestamp,
		file,
		line,
		token,
		userid,
		level,
		r.Message,
	)

	// Write to output
	_, err := h.writer.Write([]byte(logMsg))
	return err
}

// WithAttrs returns a new handler with the given attributes added
func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &customHandler{
		writer: h.writer,
		attrs:  make([]slog.Attr, len(h.attrs)+len(attrs)),
		groups: make([]string, len(h.groups)),
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.attrs[len(h.attrs):], attrs)
	copy(newHandler.groups, h.groups)
	return newHandler
}

// WithGroup returns a new handler with the given group added
func (h *customHandler) WithGroup(name string) slog.Handler {
	newHandler := &customHandler{
		writer: h.writer,
		attrs:  make([]slog.Attr, len(h.attrs)),
		groups: make([]string, len(h.groups)+1),
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.groups, h.groups)
	newHandler.groups[len(h.groups)] = name
	return newHandler
}

// Request-scoped logger
// ------------------------------------------------------------

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

// extractSessionInfoFromRequest extracts token and userid from the request
func extractSessionInfoFromRequest(r *http.Request) (token string, userid string) {
	token = "anon"
	userid = "anon"

	if r == nil {
		return
	}

	// Get session from request context
	sess := session.GetRequestSession(r)
	if sess == nil {
		return
	}

	// Extract token (key) from session
	if key, ok := sess.Get("key"); ok {
		if keyStr, ok := key.(string); ok && len(keyStr) >= 6 {
			// Get last 6 characters of token
			token = keyStr[len(keyStr)-6:]
		} else if keyStr, ok := key.(string); ok {
			// If token is shorter than 6 chars, use as is
			token = keyStr
		}
	}

	// Extract userid from session
	if uid, ok := sess.UID(); ok {
		userid = strconv.FormatInt(uid, 10)
	}

	return
}

// Logger interface methods using slog
// ------------------------------------------------------------

// Info logs an informational message
func (l *logger) Info(args ...any) {
	msg := fmt.Sprint(args...)
	l.slogger.InfoContext(l.ctx, msg)
}

// Infof logs a formatted informational message
func (l *logger) Infof(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.slogger.InfoContext(l.ctx, msg)
}

// Error logs an error message
func (l *logger) Error(args ...any) {
	msg := fmt.Sprint(args...)
	l.slogger.ErrorContext(l.ctx, msg)
}

// Errorf logs a formatted error message
func (l *logger) Errorf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.slogger.ErrorContext(l.ctx, msg)
}

// Debug logs a debug message
func (l *logger) Debug(args ...any) {
	msg := fmt.Sprint(args...)
	l.slogger.DebugContext(l.ctx, msg)
}

// Debugf logs a formatted debug message
func (l *logger) Debugf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.slogger.DebugContext(l.ctx, msg)
}

// Warn logs a warning message
func (l *logger) Warn(args ...any) {
	msg := fmt.Sprint(args...)
	l.slogger.WarnContext(l.ctx, msg)
}

// Warnf logs a formatted warning message
func (l *logger) Warnf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	l.slogger.WarnContext(l.ctx, msg)
}

// Package-level logger functions (for use without request context)
// ------------------------------------------------------------

var (
	defaultHandler = newCustomHandler(os.Stdout)
	defaultSlogger = slog.New(defaultHandler)
	defaultCtx     = withSessionInfo(context.Background(), "anon", "anon")
)

// Info logs an informational message using the default logger
func Info(args ...any) {
	msg := fmt.Sprint(args...)
	defaultSlogger.InfoContext(defaultCtx, msg)
}

// Infof logs a formatted informational message using the default logger
func Infof(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	defaultSlogger.InfoContext(defaultCtx, msg)
}

// Error logs an error message using the default logger
func Error(args ...any) {
	msg := fmt.Sprint(args...)
	defaultSlogger.ErrorContext(defaultCtx, msg)
}

// Errorf logs a formatted error message using the default logger
func Errorf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	defaultSlogger.ErrorContext(defaultCtx, msg)
}

// Debug logs a debug message using the default logger
func Debug(args ...any) {
	msg := fmt.Sprint(args...)
	defaultSlogger.DebugContext(defaultCtx, msg)
}

// Debugf logs a formatted debug message using the default logger
func Debugf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	defaultSlogger.DebugContext(defaultCtx, msg)
}

// Warn logs a warning message using the default logger
func Warn(args ...any) {
	msg := fmt.Sprint(args...)
	defaultSlogger.WarnContext(defaultCtx, msg)
}

// Warnf logs a formatted warning message using the default logger
func Warnf(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)
	defaultSlogger.WarnContext(defaultCtx, msg)
}
