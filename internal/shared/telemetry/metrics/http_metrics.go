package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("math-srv/metrics")

	// HTTP metrics
	httpRequestCounter      metric.Int64Counter
	httpRequestDuration     metric.Float64Histogram
	httpActiveRequests      metric.Int64UpDownCounter
	httpResponseSizeCounter metric.Int64Counter
)

func init() {
	var err error

	httpRequestCounter, err = meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		panic(err)
	}

	httpRequestDuration, err = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("HTTP request duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic(err)
	}

	httpActiveRequests, err = meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		panic(err)
	}

	httpResponseSizeCounter, err = meter.Int64Counter(
		"http.server.response.size",
		metric.WithDescription("Total size of HTTP responses in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		panic(err)
	}
}

// RecordHTTPRequest records metrics for an HTTP request
func RecordHTTPRequest(method, route string, statusCode int, duration time.Duration, responseSize int64) {
	ctx := context.Background()

	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", route),
		attribute.Int("http.status_code", statusCode),
	}

	// Record request count
	httpRequestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Record request duration in milliseconds
	httpRequestDuration.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))

	// Record response size
	if responseSize > 0 {
		httpResponseSizeCounter.Add(ctx, responseSize, metric.WithAttributes(attrs...))
	}
}

// IncrementActiveRequests increments the active requests counter
func IncrementActiveRequests() {
	httpActiveRequests.Add(context.Background(), 1)
}

// DecrementActiveRequests decrements the active requests counter
func DecrementActiveRequests() {
	httpActiveRequests.Add(context.Background(), -1)
}
