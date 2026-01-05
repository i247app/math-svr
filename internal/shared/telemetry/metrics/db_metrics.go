package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	// Database metrics
	dbQueryCounter      metric.Int64Counter
	dbQueryDuration     metric.Float64Histogram
	dbActiveConnections metric.Int64UpDownCounter
)

func init() {
	var err error

	dbQueryCounter, err = meter.Int64Counter(
		"db.query.count",
		metric.WithDescription("Total number of database queries"),
		metric.WithUnit("{query}"),
	)
	if err != nil {
		panic(err)
	}

	dbQueryDuration, err = meter.Float64Histogram(
		"db.query.duration",
		metric.WithDescription("Database query duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic(err)
	}

	dbActiveConnections, err = meter.Int64UpDownCounter(
		"db.active_connections",
		metric.WithDescription("Number of active database connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		panic(err)
	}
}

// RecordDBQuery records metrics for a database query
func RecordDBQuery(operation string, duration time.Duration, success bool) {
	ctx := context.Background()

	status := "success"
	if !success {
		status = "error"
	}

	attrs := []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.status", status),
	}

	// Record query count
	dbQueryCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Record query duration in milliseconds
	dbQueryDuration.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))
}

// IncrementActiveConnections increments the active connections counter
func IncrementActiveConnections() {
	dbActiveConnections.Add(context.Background(), 1)
}

// DecrementActiveConnections decrements the active connections counter
func DecrementActiveConnections() {
	dbActiveConnections.Add(context.Background(), -1)
}
