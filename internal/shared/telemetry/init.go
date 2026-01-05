package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"math-ai.com/math-ai/internal/shared/config"
)

// Providers holds the OpenTelemetry providers and shutdown functions
type Providers struct {
	TracerProvider *trace.TracerProvider
	MeterProvider  *metric.MeterProvider
	MetricsServer  *http.Server
	shutdownFuncs  []func(context.Context) error
}

// Initialize sets up OpenTelemetry tracing and metrics
func Initialize(ctx context.Context, cfg *config.OTelConfig) (*Providers, error) {
	res, err := newResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	providers := &Providers{
		shutdownFuncs: make([]func(context.Context) error, 0),
	}

	// Initialize tracing if enabled
	if cfg.EnableTracing {
		tp, shutdown, err := initTracer(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tracer: %w", err)
		}
		providers.TracerProvider = tp
		providers.shutdownFuncs = append(providers.shutdownFuncs, shutdown)
		otel.SetTracerProvider(tp)
	}

	// Initialize metrics if enabled
	if cfg.EnableMetrics {
		mp, shutdown, err := initMetrics(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
		providers.MeterProvider = mp
		providers.shutdownFuncs = append(providers.shutdownFuncs, shutdown)
		otel.SetMeterProvider(mp)

		// Start Prometheus metrics HTTP server
		providers.MetricsServer = StartMetricsServer(cfg.PrometheusPort)
	}

	// Set global propagator for W3C Trace Context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return providers, nil
}

// initTracer initializes the trace provider with OTLP HTTP exporter
func initTracer(ctx context.Context, cfg *config.OTelConfig, res *resource.Resource) (*trace.TracerProvider, func(context.Context) error, error) {
	// Create OTLP HTTP exporter for Jaeger
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Configure sampler based on sample rate
	sampler := trace.ParentBased(trace.TraceIDRatioBased(cfg.TraceSampleRate))

	// Create trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(sampler),
	)

	return tp, tp.Shutdown, nil
}

// initMetrics initializes the metric provider with Prometheus exporter
func initMetrics(ctx context.Context, cfg *config.OTelConfig, res *resource.Resource) (*metric.MeterProvider, func(context.Context) error, error) {
	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create meter provider
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(exporter),
	)

	return mp, mp.Shutdown, nil
}

// Shutdown gracefully shuts down all OpenTelemetry providers
func (p *Providers) Shutdown(ctx context.Context) error {
	// Stop metrics server first
	if p.MetricsServer != nil {
		if err := StopMetricsServer(p.MetricsServer); err != nil {
			return fmt.Errorf("failed to stop metrics server: %w", err)
		}
	}

	// Shutdown providers
	for _, shutdown := range p.shutdownFuncs {
		if err := shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown OpenTelemetry provider: %w", err)
		}
	}
	return nil
}
