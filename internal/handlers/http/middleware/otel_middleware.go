package middleware

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"math-ai.com/math-ai/internal/shared/metadata"
	"math-ai.com/math-ai/internal/shared/telemetry/metrics"
)

var httpTracer = otel.Tracer("math-srv/http")

// OtelMiddleware creates OpenTelemetry instrumentation middleware for HTTP requests
func OtelMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from headers (W3C Trace Context)
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Create span name from method and path
			spanName := r.Method + " " + r.URL.Path
			ctx, span := httpTracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPMethod(r.Method),
					semconv.HTTPRoute(r.URL.Path),
					semconv.HTTPScheme(r.URL.Scheme),
					semconv.HTTPTarget(r.URL.RequestURI()),
					semconv.NetHostName(r.Host),
					semconv.UserAgentOriginal(r.UserAgent()),
				),
			)
			defer span.End()

			// Add request metadata as span attributes if available
			addMetadataAttributes(ctx, span)

			// Wrap response writer to capture status code and response size
			wrw := &otelResponseWriter{
				ResponseWriter: w,
				statusCode:     200, // default to 200
				bytesWritten:   0,
			}

			// Record start time for metrics
			start := time.Now()

			// Increment active requests counter
			metrics.IncrementActiveRequests()
			defer metrics.DecrementActiveRequests()

			// Handle panic recovery with span recording
			defer func() {
				if rec := recover(); rec != nil {
					// Record error in span
					err := fmt.Errorf("panic: %v", rec)
					span.RecordError(err)
					span.SetStatus(codes.Error, "panic recovered")

					// Re-panic to let RecoveryMiddleware handle it
					panic(rec)
				}
			}()

			// Continue request chain with traced context
			next.ServeHTTP(wrw, r.WithContext(ctx))

			// Record final span attributes after request completes
			duration := time.Since(start)
			span.SetAttributes(
				semconv.HTTPStatusCode(wrw.statusCode),
				attribute.Int64("http.response.size", wrw.bytesWritten),
				attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
			)

			// Set span status based on HTTP status code
			if wrw.statusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(wrw.statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}

			// Record metrics
			metrics.RecordHTTPRequest(r.Method, r.URL.Path, wrw.statusCode, duration, wrw.bytesWritten)
		})
	}
}

// addMetadataAttributes extracts metadata from context and adds it to the span
func addMetadataAttributes(ctx context.Context, span trace.Span) {
	md := metadata.FromContext(ctx)
	if md != nil && !md.IsEmpty() {
		span.SetAttributes(
			attribute.String("app.trace_id", md.TraceID),
			attribute.String("app.request_id", md.RequestID),
			attribute.String("client.platform", md.ClientInfo.Platform),
			attribute.String("client.app_version", md.ClientInfo.AppVersion),
			attribute.String("client.device_model", md.ClientInfo.DeviceModel),
			attribute.String("client.os_version", md.ClientInfo.OSVersion),
			attribute.String("client.device_id", md.ClientInfo.DeviceID),
			attribute.String("user.locale", md.UserContext.Locale),
			attribute.String("user.language", md.UserContext.Language),
			attribute.String("user.timezone", md.UserContext.Timezone),
		)
	}
}

// otelResponseWriter wraps http.ResponseWriter to capture status code and bytes written
type otelResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	wroteHeader  bool
}

func (w *otelResponseWriter) WriteHeader(statusCode int) {
	if !w.wroteHeader {
		w.statusCode = statusCode
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *otelResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// Implement http.Flusher interface if underlying ResponseWriter supports it
func (w *otelResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Implement http.Hijacker interface if underlying ResponseWriter supports it
func (w *otelResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("response writer does not support hijacking")
}
