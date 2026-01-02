package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
)

// Custom slog.Handler implementation
// ------------------------------------------------------------

// customHandler implements slog.Handler with our custom format
type customHandler struct {
	writer          io.Writer
	attrs           []slog.Attr
	groups          []string
	backgroundColor BackgroundColor
}

// newCustomHandler creates a new custom handler that writes to the given writer
func newCustomHandler(w io.Writer) *customHandler {
	return newCustomHandlerWithBgColor(w, BgNone)
}

// newCustomHandlerWithBgColor creates a new custom handler with a specific background color
func newCustomHandlerWithBgColor(w io.Writer, bgColor BackgroundColor) *customHandler {
	return &customHandler{
		writer:          w,
		attrs:           make([]slog.Attr, 0),
		groups:          make([]string, 0),
		backgroundColor: bgColor,
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
	token, userid, route := extractSessionInfo(ctx)

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
	timestamp := r.Time.Format("2006/01/02 15:04:05")

	// Format level
	level := r.Level.String()

	// Build the log message WITHOUT the trailing newline
	logMsg := fmt.Sprintf("%s [%s] [%s] [%s] %s:%d %s: %s",
		timestamp,
		token,
		userid,
		route,
		file,
		line,
		level,
		r.Message,
	)

	// Check for background color in context, fallback to handler's default
	bgColor := GetBackgroundColor(ctx)
	// if bgColor == BgNone {
	// 	bgColor = h.backgroundColor
	// }

	// Apply background color to text only (not the newline)
	logMsg = applyBackgroundColor(logMsg, bgColor)

	// Add newline AFTER applying background color
	logMsg += "\n"

	// Write to output
	_, err := h.writer.Write([]byte(logMsg))
	return err
}

// WithAttrs returns a new handler with the given attributes added
func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &customHandler{
		writer:          h.writer,
		attrs:           make([]slog.Attr, len(h.attrs)+len(attrs)),
		groups:          make([]string, len(h.groups)),
		backgroundColor: h.backgroundColor,
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.attrs[len(h.attrs):], attrs)
	copy(newHandler.groups, h.groups)
	return newHandler
}

// WithGroup returns a new handler with the given group added
func (h *customHandler) WithGroup(name string) slog.Handler {
	newHandler := &customHandler{
		writer:          h.writer,
		attrs:           make([]slog.Attr, len(h.attrs)),
		groups:          make([]string, len(h.groups)+1),
		backgroundColor: h.backgroundColor,
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.groups, h.groups)
	newHandler.groups[len(h.groups)] = name
	return newHandler
}
