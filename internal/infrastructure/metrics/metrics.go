// Package metrics owns the application's Prometheus registry and the HTTP
// RED (Rate / Errors / Duration) collectors. One *Metrics is built at boot
// (bootstrap.SetupResource) and shared read-only afterwards.
//
// A nil *Metrics means metrics are disabled (OBS_METRICS_ENABLED=false):
// every method is a safe no-op, so callers never need to nil-check inline.
package metrics

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Metrics bundles the registry (exposed at /metrics) and the HTTP collectors.
type Metrics struct {
	Registry *prometheus.Registry

	httpRequests *prometheus.CounterVec
	httpDuration *prometheus.HistogramVec
	httpInFlight prometheus.Gauge
}

// New builds a private registry (not the global default, to avoid picking up
// collectors registered by dependencies), wires the HTTP + Go runtime +
// process collectors, and stamps a build_info gauge.
func New(serviceName, version string) *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		Registry: reg,
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed, by method, route and status.",
		}, []string{"method", "route", "status"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds, by method, route and status.",
			Buckets: prometheus.DefBuckets, // 5ms … 10s
		}, []string{"method", "route", "status"}),
		httpInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "HTTP requests currently being served.",
		}),
	}

	reg.MustRegister(
		m.httpRequests,
		m.httpDuration,
		m.httpInFlight,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "math_svr_build_info",
		Help: "Build/version info; the value is always 1.",
	}, []string{"service", "version"})
	buildInfo.WithLabelValues(serviceName, version).Set(1)
	reg.MustRegister(buildInfo)

	return m
}

// ObserveHTTP records one finished HTTP request. No-op on a nil receiver.
//
// When traceID is non-empty it is attached to the latency histogram as a
// Prometheus exemplar, letting Grafana jump from a point on the latency graph
// straight to that request's trace in Tempo. Exemplars are stored out-of-band
// (a few per bucket), so the high-cardinality trace_id does NOT create new
// time series — this is safe and intended.
func (m *Metrics) ObserveHTTP(method, route string, status int, seconds float64, traceID string) {
	if m == nil {
		return
	}
	code := strconv.Itoa(status)
	m.httpRequests.WithLabelValues(method, route, code).Inc()

	obs := m.httpDuration.WithLabelValues(method, route, code)
	if traceID != "" {
		if eo, ok := obs.(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(seconds, prometheus.Labels{"trace_id": traceID})
			return
		}
	}
	obs.Observe(seconds)
}

// RegisterSocketConnections wires a live gauge of active WebSocket connections,
// sampled from fn at scrape time (WebSocket connections are long-lived, so a
// pull-based gauge fits better than hot-path counters). No-op on a nil
// receiver, so it is safe to call unconditionally when metrics are disabled.
func (m *Metrics) RegisterSocketConnections(fn func() int) {
	if m == nil {
		return
	}
	m.Registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "socket_active_connections",
			Help: "Active realtime WebSocket connections.",
		},
		func() float64 { return float64(fn()) },
	))
}

// IncInFlight / DecInFlight bracket a request. No-op on a nil receiver.
func (m *Metrics) IncInFlight() {
	if m != nil {
		m.httpInFlight.Inc()
	}
}

func (m *Metrics) DecInFlight() {
	if m != nil {
		m.httpInFlight.Dec()
	}
}

// NormalizeRoute collapses high-cardinality path segments (numeric IDs, long
// token/uuid-ish strings) into ":id" so the `route` metric label stays
// bounded. The metrics middleware runs before gex's mux matches the route, so
// the matched pattern (r.Pattern) is not yet available — this is a
// best-effort reconstruction of the route template from the raw path.
//
// Examples:
//
//	/users/42            → /users/:id
//	/users/create        → /users/create        (unchanged)
//	/classrooms/9f3a…len  → /classrooms/:id      (long token with a digit)
func NormalizeRoute(path string) string {
	if path == "" || path == "/" {
		return path
	}
	segs := strings.Split(path, "/")
	for i, s := range segs {
		if s == "" {
			continue
		}
		if looksLikeID(s) {
			segs[i] = ":id"
		}
	}
	return strings.Join(segs, "/")
}

// looksLikeID reports whether a single path segment is likely a volatile
// identifier rather than a fixed route word. Pure-digit segments always
// qualify; otherwise only long (≥20 char) tokens that contain a digit do —
// this catches uuids/hashes/opaque tokens while leaving normal words alone.
func looksLikeID(s string) bool {
	allDigits := true
	for _, r := range s {
		if r < '0' || r > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		return true
	}
	if len(s) >= 20 {
		for _, r := range s {
			if r >= '0' && r <= '9' {
				return true
			}
		}
	}
	return false
}
