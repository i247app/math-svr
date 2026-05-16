// Package logger provides a per-request, slog-backed logger that emits
// binbase-svr's house log format:
//
//	2026/04/27 08:20:33 [6MEuCw] [623119] [POST /lookups/sync] msg key=val ...
//
// Middleware (internal/bootstrap/middleware.LoggerMiddleware) constructs one
// AppLogger per HTTP request and binds it to the request context. Handlers
// and services retrieve it with logger.From(ctx); background work that
// outlives the request copies the logger via NewBackground.
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

	"math-ai.com/math-ai/internal/shared/colors"
)

// AppLogger is a per-request structured logger. It wraps a *slog.Logger
// whose handler emits the binbase-svr prefix format. Holding the request lets
// every log line written during the request show the same [METHOD path]
// without callers re-passing it.
//
// Concurrency: safe for use from any goroutine spawned within the request's
// lifetime. The underlying handler serialises Writes.
type AppLogger struct {
	sLogger         *slog.Logger
	request         *http.Request
	outFile         *os.File
	ctx             context.Context
	backgroundColor colors.Color
	colorEnabled    bool
}

// New constructs an AppLogger bound to ctx and r. r may be nil for
// background work — the request-line slot then renders as "[-]".
func New(ctx context.Context, r *http.Request, opts Options) *AppLogger {
	out := io.Writer(os.Stdout)
	if opts.Output != nil {
		out = opts.Output
	}

	h := newPrefixHandler(out, r, opts.Level, opts.BackgroundColor, opts.WithColor)
	al := &AppLogger{
		sLogger:         slog.New(h),
		request:         r,
		ctx:             ctx,
		backgroundColor: opts.BackgroundColor,
		colorEnabled:    opts.WithColor,
	}
	if f, ok := out.(*os.File); ok {
		al.outFile = f
	}
	return al
}

// NewBackground returns a logger detached from any request, suitable for
// goroutines spawned for fire-and-forget work (NATS publishers, cron jobs).
// The format degrades to [anon] [anon] [-] unless ctx already carries
// uid / token suffix.
func NewBackground(ctx context.Context, opts Options) *AppLogger {
	return New(ctx, nil, opts)
}

// Slog exposes the underlying *slog.Logger for advanced callers (e.g.
// passing to libraries that accept slog directly). Most code should use the
// Info / Warn / Error / Debug shortcuts on AppLogger.
func (l *AppLogger) Slog() *slog.Logger { return l.sLogger }

// Request returns the bound *http.Request, or nil for background loggers.
func (l *AppLogger) Request() *http.Request { return l.request }

// Context returns the bound context. Useful in helpers that need the same
// trace id / uid binding the logger uses.
func (l *AppLogger) Context() context.Context { return l.ctx }

// ColorEnabled reports whether ANSI escape codes embedded in log messages
// will reach a color-capable terminal. Helpers that emit pre-colored
// segments (e.g. SQL trace lines with the query in cyan and args in
// yellow) consult this so their log files stay free of escape codes when
// the logger is configured for plain output.
func (l *AppLogger) ColorEnabled() bool { return l.colorEnabled }

// Info / Warn / Error / Debug forward to slog at the matching level.
// args follow slog's alternating-key/value convention; pass slog.Attr-
// typed values when you want explicit structured fields. The handler's
// rendered file:line attribution points at the caller of these methods.
func (l *AppLogger) Info(msg string, args ...any)  { l.emit(slog.LevelInfo, msg, args) }
func (l *AppLogger) Warn(msg string, args ...any)  { l.emit(slog.LevelWarn, msg, args) }
func (l *AppLogger) Error(msg string, args ...any) { l.emit(slog.LevelError, msg, args) }
func (l *AppLogger) Debug(msg string, args ...any) { l.emit(slog.LevelDebug, msg, args) }

// Infof / Warnf / Errorf / Debugf are printf-style shortcuts: they format
// the message via fmt.Sprintf and emit it as a single message string. No
// structured attributes are attached — the formatted string is the whole
// payload. Use these when you're porting log.Printf calls or when the line
// is purely human-readable narration; prefer the slog-style key/value
// form (Info / Warn / Error / Debug) for fields you'll grep or aggregate.
func (l *AppLogger) Infof(format string, args ...any) {
	l.emit(slog.LevelInfo, fmt.Sprintf(format, args...), nil)
}
func (l *AppLogger) Warnf(format string, args ...any) {
	l.emit(slog.LevelWarn, fmt.Sprintf(format, args...), nil)
}
func (l *AppLogger) Errorf(format string, args ...any) {
	l.emit(slog.LevelError, fmt.Sprintf(format, args...), nil)
}
func (l *AppLogger) Debugf(format string, args ...any) {
	l.emit(slog.LevelDebug, fmt.Sprintf(format, args...), nil)
}

// callerSkip is the number of stack frames runtime.Callers must skip to
// land on user code. Frame layout under emit:
//  0. runtime.Callers
//  1. emit (this function)
//  2. Info / Warn / Error / Debug / their f-variants
//  3. user call site                       ← what we want
//
// All log methods funnel through emit, so this constant stays valid as
// long as that contract holds. Adding a new wrapper level would require
// either adjusting the skip per call path or re-using emit unchanged.
const callerSkip = 3

// emit captures the user's program counter, builds a slog.Record with it,
// and hands it to the underlying handler. Doing this here (rather than
// going through slog.Logger.Log) is what lets prefixHandler render the
// caller's file:line — slog.Logger.Log captures the PC of its own caller,
// which would always be one of our wrapper methods.
func (l *AppLogger) emit(level slog.Level, msg string, args []any) {
	if !l.sLogger.Enabled(l.ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(callerSkip, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = l.sLogger.Handler().Handle(l.ctx, r)
}

// With returns a new AppLogger that emits attrs on every record. Cheap —
// reuses the same handler chain via slog.Logger.With.
func (l *AppLogger) With(args ...any) *AppLogger {
	return &AppLogger{
		sLogger:         l.sLogger.With(args...),
		request:         l.request,
		outFile:         l.outFile,
		ctx:             l.ctx,
		backgroundColor: l.backgroundColor,
		colorEnabled:    l.colorEnabled,
	}
}

// Close releases resources owned by the logger. No-op when the output is
// not a process-managed file. Stdout / stderr are intentionally NOT closed.
func (l *AppLogger) Close() error {
	if l.outFile == nil {
		return nil
	}
	if l.outFile == os.Stdout || l.outFile == os.Stderr {
		return nil
	}
	return l.outFile.Close()
}
