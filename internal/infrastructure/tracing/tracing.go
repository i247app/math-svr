// Package tracing wires OpenTelemetry distributed tracing for math-svr.
//
// Init installs a global TracerProvider that batches spans to an OTLP/HTTP
// endpoint (Tempo). When tracing is disabled the global provider stays the
// SDK default (a no-op), so every otel.Tracer(...).Start(...) call elsewhere
// in the codebase is a cheap no-op — instrumentation code never has to branch
// on "is tracing enabled".
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const tracerName = "math-svr"

// Tracer returns the app's tracer. Safe to call before Init — it returns a
// no-op tracer until the global provider is installed.
func Tracer() trace.Tracer { return otel.Tracer(tracerName) }

// Init installs the global TracerProvider + W3C propagator from cfg and
// returns a shutdown func that flushes pending spans on exit. When
// cfg.TracingEnabled is false it wires nothing and returns a no-op shutdown,
// leaving the global no-op tracer in place.
func Init(ctx context.Context, cfg config.ObservabilityConfig) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }
	if !cfg.TracingEnabled {
		return noop, nil
	}

	opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.OTLPEndpoint)}
	if cfg.OTLPInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return noop, fmt.Errorf("tracing: create OTLP exporter: %w", err)
	}

	// resource.Default() carries SDK + telemetry attrs and a schema URL;
	// NewSchemaless adds our identity attrs without a (conflicting) schema.
	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", cfg.ServiceVersion),
		attribute.String("deployment.environment", cfg.Environment),
	))
	if err != nil {
		return noop, fmt.Errorf("tracing: build resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		// ParentBased: honour an upstream sampling decision; otherwise sample
		// at the configured ratio (1.0 = everything, for local dev).
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSampleRatio))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Bridge trace ids into structured logs → Grafana Loki↔Tempo correlation.
	logger.TraceContextFn = TraceIDs

	return func(shutdownCtx context.Context) error {
		logger.TraceContextFn = nil
		return tp.Shutdown(shutdownCtx)
	}, nil
}

// TraceIDs extracts the active trace/span ids from ctx as hex strings, or
// ("","") when ctx carries no valid span. Wired into logger.TraceContextFn by
// Init so every JSON log line during a traced request carries its trace_id.
func TraceIDs(ctx context.Context) (traceID, spanID string) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return "", ""
	}
	return sc.TraceID().String(), sc.SpanID().String()
}
