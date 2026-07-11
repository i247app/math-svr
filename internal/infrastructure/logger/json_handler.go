package logger

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	sctx "math-ai.com/math-ai/internal/shared/context"
)

// TraceContextFn, when set, extracts the active trace/span ids from ctx so
// JSON logs can carry them — this is what powers Grafana's log↔trace
// correlation (Loki derived field → Tempo). It is wired in Phase 3
// (tracing.Init) to avoid a hard OpenTelemetry dependency in the logger
// package. Nil → no trace fields are emitted (Phase 2 state).
var TraceContextFn func(ctx context.Context) (traceID, spanID string)

// jsonHandler emits one JSON object per record, carrying the same contextual
// fields as prefixHandler (token, uid, method, path, caller) plus
// trace_id/span_id when available. Keys are stable so the Alloy pipeline
// (stage.json extracting `level` + `trace_id`) keeps working.
//
// One handler is built per AppLogger (per request); Handle serialises writes
// via mu so concurrent goroutines within a request don't interleave lines.
type jsonHandler struct {
	out     io.Writer
	mu      *sync.Mutex
	level   slog.Leveler
	request *http.Request

	attrs       []slog.Attr
	groupPrefix string
}

func newJSONHandler(out io.Writer, r *http.Request, level slog.Leveler) *jsonHandler {
	if level == nil {
		level = slog.LevelInfo
	}
	return &jsonHandler{
		out:     out,
		mu:      &sync.Mutex{},
		level:   level,
		request: r,
	}
}

func (h *jsonHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level.Level()
}

func (h *jsonHandler) Handle(ctx context.Context, rec slog.Record) error {
	m := make(map[string]any, 12)
	m["time"] = rec.Time.UTC().Format(time.RFC3339Nano)
	m["level"] = levelString(rec.Level)
	m["msg"] = rec.Message

	if rid := sctx.RequestID(ctx); rid != 0 {
		m["request_id"] = rid
	}
	if tok := tokenSuffixOr(ctx); tok != "" && tok != anonToken {
		m["token"] = tok
	}
	if uid := sctx.UserID(ctx); uid != 0 {
		m["uid"] = uid
	}
	if h.request != nil {
		m["method"] = h.request.Method
		if h.request.URL != nil {
			m["path"] = h.request.URL.Path
		}
	}
	if TraceContextFn != nil {
		if tid, sid := TraceContextFn(ctx); tid != "" {
			m["trace_id"] = tid
			if sid != "" {
				m["span_id"] = sid
			}
		}
	}
	if frame, ok := pcFrame(rec.PC); ok {
		m["caller"] = filepath.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
	}

	for _, a := range h.attrs {
		addJSONAttr(m, h.groupPrefix, a)
	}
	rec.Attrs(func(a slog.Attr) bool {
		addJSONAttr(m, h.groupPrefix, a)
		return true
	})

	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}
	buf = append(buf, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err = h.out.Write(buf)
	return err
}

func (h *jsonHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	cp := *h
	cp.attrs = make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	cp.attrs = append(cp.attrs, h.attrs...)
	cp.attrs = append(cp.attrs, attrs...)
	return &cp
}

func (h *jsonHandler) WithGroup(name string) slog.Handler {
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

// addJSONAttr writes one attr into m, resolving LogValuers and flattening
// groups with a dotted prefix.
func addJSONAttr(m map[string]any, prefix string, a slog.Attr) {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return
	}
	if a.Value.Kind() == slog.KindGroup {
		grp := a.Value.Group()
		if len(grp) == 0 {
			return
		}
		np := prefix + a.Key + "."
		if a.Key == "" {
			np = prefix
		}
		for _, ga := range grp {
			addJSONAttr(m, np, ga)
		}
		return
	}
	m[prefix+a.Key] = a.Value.Any()
}

func levelString(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "error"
	case l >= slog.LevelWarn:
		return "warn"
	case l >= slog.LevelInfo:
		return "info"
	default:
		return "debug"
	}
}
