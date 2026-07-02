package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"math-ai.com/math-ai/internal/infrastructure/httproute"
)

const httpTracerName = "math-svr/http"

// TracingMiddleware starts a server span per request, continuing any upstream
// trace carried in the W3C `traceparent` header. Register it just INSIDE
// MetricsMiddleware and BEFORE LoggerMiddleware, so the logger stamps the
// trace_id onto every line emitted during the request.
//
// enabled=false → pass-through (no span, no propagation cost). This mirrors
// the metrics middleware's nil-guard so OBS_TRACING_ENABLED is honoured at
// one place with no per-request branching downstream.
//
// The span is named "<METHOD> <route>" using the normalized route (ids
// collapsed to :id) to keep span names / metrics low-cardinality.
func TracingMiddleware(enabled bool, classifier *httproute.Classifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !enabled {
			return next
		}
		propagator := otel.GetTextMapPropagator()
		tracer := otel.Tracer(httpTracerName)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			route := routeLabel(classifier, r)
			ctx, span := tracer.Start(ctx, r.Method+" "+route,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.request.method", r.Method),
					attribute.String("http.route", route),
					attribute.String("url.path", r.URL.Path),
					attribute.String("client.address", r.RemoteAddr),
				),
			)
			defer span.End()

			// Publish the trace id back to the outer metrics middleware so it
			// can attach a latency-histogram exemplar (metric → trace jump).
			if sc := span.SpanContext(); sc.IsValid() {
				setTraceIDHolder(ctx, sc.TraceID().String())
			}

			sw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))

			span.SetAttributes(attribute.Int("http.response.status_code", sw.status))
			if sw.status >= 500 {
				span.SetStatus(codes.Error, http.StatusText(sw.status))
			}
		})
	}
}
