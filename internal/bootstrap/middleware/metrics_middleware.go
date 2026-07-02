package middleware

import (
	"context"
	"net/http"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/httproute"
	"math-ai.com/math-ai/internal/infrastructure/metrics"
)

// routeLabel resolves the low-cardinality route label for r. It prefers the
// registered route template from the classifier (unknown paths → "unmatched")
// and falls back to a heuristic path normalization when no classifier is
// wired (e.g. in unit tests).
func routeLabel(c *httproute.Classifier, r *http.Request) string {
	if c != nil {
		return c.Route(r)
	}
	return metrics.NormalizeRoute(r.URL.Path)
}

// traceIDHolder is a mutable slot passed down the request context so the
// (outer) metrics middleware can read the trace_id that the (inner) tracing
// middleware discovers — enabling latency-histogram exemplars. Kept in this
// package so both middlewares share it without an import cycle.
type traceIDHolder struct{ id string }

type traceHolderKeyType struct{}

var traceHolderKey traceHolderKeyType

// setTraceIDHolder is called by the tracing middleware to publish the active
// trace id back to the metrics middleware. No-op when no holder is present
// (metrics disabled).
func setTraceIDHolder(ctx context.Context, traceID string) {
	if h, ok := ctx.Value(traceHolderKey).(*traceIDHolder); ok {
		h.id = traceID
	}
}

// MetricsMiddleware records RED metrics (rate, errors, duration) for every
// HTTP request. Register it as the OUTERMOST middleware so the measured
// duration covers the whole request, including all inner middleware.
//
// A nil *metrics.Metrics disables the middleware entirely (pass-through) —
// this is how OBS_METRICS_ENABLED=false is honoured without conditionals at
// the registration site.
//
// NOTE on the `status` label: this project's response.WriteJson always emits
// HTTP 200 (the semantic code lives in the JSON body's `mstatus`). So the
// status label will almost always be "200"; treat HTTP-level errors here as
// transport failures/panics, and use body-level mstatus for business errors.
func MetricsMiddleware(m *metrics.Metrics, classifier *httproute.Classifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if m == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't let Prometheus scraping its own endpoint inflate counters.
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			m.IncInFlight()
			defer m.DecInFlight()

			// Provide a slot the tracing middleware (inner) fills with the
			// trace id, so we can attach it as a histogram exemplar below.
			holder := &traceIDHolder{}
			ctx := context.WithValue(r.Context(), traceHolderKey, holder)

			sw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))

			route := routeLabel(classifier, r)
			m.ObserveHTTP(r.Method, route, sw.status, time.Since(start).Seconds(), holder.id)
		})
	}
}

// statusRecorder captures the HTTP status code written by inner handlers.
// It implements Unwrap so http.ResponseController (and thus Flusher/Hijacker
// used by the gzip/streaming paths) can reach the underlying writer.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if !s.wroteHeader {
		s.status = code
		s.wroteHeader = true
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if !s.wroteHeader {
		s.wroteHeader = true
	}
	return s.ResponseWriter.Write(b)
}

func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }
