package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"math-ai.com/math-ai/internal/shared/colors"
	sctx "math-ai.com/math-ai/internal/shared/context"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	anonToken  = "anon"
	anonUID    = "anon"
	noRequest  = "-"
	timeFormat = "2006/01/02 15:04:05"
)

// prefixHandler is the slog.Handler that produces the binbase-svr log line:
//
//	YYYY/MM/DD HH:MM:SS [<token6>] [<uid>] [<METHOD> <path>] msg key=val ...
//
// One handler is constructed per AppLogger (i.e. per request) so it can hold
// the request reference. Concurrent Handle calls are serialised via mu to
// keep multi-line records (panic stacks, dumps) intact in the output stream.
type prefixHandler struct {
	out             io.Writer
	mu              *sync.Mutex
	level           slog.Leveler
	request         *http.Request
	backgroundColor colors.Color
	withColor       bool

	// attrs accumulated by WithAttrs; emitted on every record.
	attrs []slog.Attr
	// group prefix accumulated by WithGroup; applied to attr keys.
	groupPrefix string
}

func newPrefixHandler(out io.Writer, r *http.Request, level slog.Leveler, bg colors.Color, withColor bool) *prefixHandler {
	if level == nil {
		level = slog.LevelInfo
	}
	return &prefixHandler{
		out:             out,
		mu:              &sync.Mutex{},
		level:           level,
		request:         r,
		backgroundColor: bg,
		withColor:       withColor,
	}
}

func (h *prefixHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level.Level()
}

func (h *prefixHandler) Handle(ctx context.Context, rec slog.Record) error {
	var b strings.Builder
	b.Grow(96 + len(rec.Message))

	b.WriteString(rec.Time.UTC().Format(timeFormat))
	b.WriteByte(' ')

	b.WriteByte('[')
	b.WriteString(tokenSuffixOr(ctx))
	b.WriteString("] [")
	b.WriteString(uidOr(ctx))
	b.WriteString("] [")
	b.WriteString(h.requestLine())
	b.WriteString("] ")

	// Caller attribution. AppLogger.emit captures runtime.Callers(3, …)
	// so rec.PC points at the user's call site (handler / service / repo
	// helper); resolving it here keeps the format stable across all the
	// log methods (Info, Infof, Warn, …). When PC is zero — slog records
	// constructed without a PC — the section is silently skipped.
	if frame, ok := pcFrame(rec.PC); ok {
		b.WriteString(filepath.Base(frame.File))
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(frame.Line))
		b.WriteByte(' ')
	}

	b.WriteString(rec.Message)

	for _, a := range h.attrs {
		writeAttr(&b, h.groupPrefix, a)
	}
	rec.Attrs(func(a slog.Attr) bool {
		writeAttr(&b, h.groupPrefix, a)
		return true
	})
	b.WriteByte('\n')

	line := b.String()
	if h.withColor {
		line = colors.Colorize(levelColor(rec.Level), h.backgroundColor, line)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.out, line)
	return err
}

func (h *prefixHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	cp := *h
	cp.attrs = make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	cp.attrs = append(cp.attrs, h.attrs...)
	cp.attrs = append(cp.attrs, attrs...)
	return &cp
}

func (h *prefixHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	cp := *h
	if cp.groupPrefix == "" {
		cp.groupPrefix = name + "."
	} else {
		cp.groupPrefix = cp.groupPrefix + name + "."
	}
	return &cp
}

func (h *prefixHandler) requestLine() string {
	if h.request == nil {
		return noRequest
	}
	path := ""
	if h.request.URL != nil {
		path = h.request.URL.Path
	}
	return h.request.Method + " " + path
}

func tokenSuffixOr(ctx context.Context) string {
	if v := sctx.TokenSuffix(ctx); v != "" {
		return v
	}
	return anonToken
}

func uidOr(ctx context.Context) string {
	if v := sctx.UserID(ctx); v != 0 {
		return utils.Int64ToString(v)
	}
	return anonUID
}

func writeAttr(b *strings.Builder, prefix string, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}
	b.WriteByte(' ')
	b.WriteString(prefix)
	b.WriteString(a.Key)
	b.WriteByte('=')
	// Use slog's stringer; quote if value contains spaces.
	v := a.Value.String()
	if strings.ContainsAny(v, " \t") {
		fmt.Fprintf(b, "%q", v)
		return
	}
	b.WriteString(v)
}

// pcFrame resolves a program counter to its source frame. Returns false
// when pc is zero (e.g. a slog.Record constructed with no PC) or when the
// runtime cannot resolve it. Only the first frame is read — inlined call
// sites still report their own file:line through this path.
func pcFrame(pc uintptr) (runtime.Frame, bool) {
	if pc == 0 {
		return runtime.Frame{}, false
	}
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()
	if frame.File == "" {
		return runtime.Frame{}, false
	}
	return frame, true
}

func levelColor(lvl slog.Level) colors.Color {
	switch {
	case lvl >= slog.LevelError:
		return colors.FGRed
	case lvl >= slog.LevelWarn:
		return colors.FGYellow
	case lvl >= slog.LevelInfo:
		return colors.FGGreen
	default:
		return colors.FGCyan
	}
}
