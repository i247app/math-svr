package logger

import (
	"context"
	"errors"
	"log/slog"
)

// multiHandler fans one record out to several slog.Handlers (e.g. a text
// console handler + a JSON file handler), so a single AppLogger can write the
// same event to multiple destinations in different encodings.
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(hs ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: hs}
}

func (m *multiHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, lvl) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, rec slog.Record) error {
	var errs []error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, rec.Level) {
			continue
		}
		// Clone so a handler that retains/mutates the record can't affect the
		// next one (slog.Handler contract).
		if err := h.Handle(ctx, rec.Clone()); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		nh[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: nh}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	nh := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		nh[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: nh}
}
