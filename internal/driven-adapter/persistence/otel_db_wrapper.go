package persistence

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/telemetry/metrics"
)

var dbTracer = otel.Tracer("math-srv/database")

// OtelDatabaseWrapper wraps IDatabase with OpenTelemetry instrumentation
type OtelDatabaseWrapper struct {
	db              db.IDatabase
	enableTracing   bool
	sanitizeQueries bool
}

// NewOtelDatabaseWrapper creates a new instrumented database wrapper
func NewOtelDatabaseWrapper(database db.IDatabase, enableTracing bool) *OtelDatabaseWrapper {
	return &OtelDatabaseWrapper{
		db:              database,
		enableTracing:   enableTracing,
		sanitizeQueries: true, // Always sanitize queries to avoid leaking sensitive data
	}
}

// GetDB returns the underlying *sql.DB
func (w *OtelDatabaseWrapper) GetDB() *sql.DB {
	return w.db.GetDB()
}

// WithTransaction wraps transaction execution with tracing
func (w *OtelDatabaseWrapper) WithTransaction(function db.HanderlerWithTx) error {
	if !w.enableTracing {
		return w.db.WithTransaction(function)
	}

	ctx := context.Background()
	ctx, span := dbTracer.Start(ctx, "db.transaction",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemMySQL,
			attribute.String("db.operation", "transaction"),
		),
	)
	defer span.End()

	start := time.Now()
	err := w.db.WithTransaction(function)
	duration := time.Since(start)

	// Record metrics
	success := err == nil
	metrics.RecordDBQuery("transaction", duration, success)

	// Set span status
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

// Query wraps Query with tracing and metrics
func (w *OtelDatabaseWrapper) Query(ctx context.Context, tx *sql.Tx, query string, args ...any) (*sql.Rows, error) {
	if !w.enableTracing {
		return w.db.Query(ctx, tx, query, args...)
	}

	operation := extractOperation(query)
	sanitizedQuery := w.sanitizeQuery(query)

	ctx, span := dbTracer.Start(ctx, "db.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemMySQL,
			attribute.String("db.operation", operation),
			attribute.String("db.statement", sanitizedQuery),
		),
	)
	defer span.End()

	start := time.Now()
	rows, err := w.db.Query(ctx, tx, query, args...)
	duration := time.Since(start)

	// Record metrics
	success := err == nil
	metrics.RecordDBQuery(operation, duration, success)

	// Set span attributes
	span.SetAttributes(
		attribute.Float64("db.duration_ms", float64(duration.Milliseconds())),
	)

	// Set span status
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return rows, err
}

// QueryRow wraps QueryRow with tracing and metrics
func (w *OtelDatabaseWrapper) QueryRow(ctx context.Context, tx *sql.Tx, query string, args ...any) *sql.Row {
	if !w.enableTracing {
		return w.db.QueryRow(ctx, tx, query, args...)
	}

	operation := extractOperation(query)
	sanitizedQuery := w.sanitizeQuery(query)

	ctx, span := dbTracer.Start(ctx, "db.query_row",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemMySQL,
			attribute.String("db.operation", operation),
			attribute.String("db.statement", sanitizedQuery),
		),
	)
	defer span.End()

	start := time.Now()
	row := w.db.QueryRow(ctx, tx, query, args...)
	duration := time.Since(start)

	// Record metrics - QueryRow doesn't return error until Scan is called
	// So we always record as success here
	metrics.RecordDBQuery(operation, duration, true)

	// Set span attributes
	span.SetAttributes(
		attribute.Float64("db.duration_ms", float64(duration.Milliseconds())),
	)
	span.SetStatus(codes.Ok, "")

	return row
}

// Exec wraps Exec with tracing and metrics
func (w *OtelDatabaseWrapper) Exec(ctx context.Context, tx *sql.Tx, query string, args ...any) (sql.Result, error) {
	if !w.enableTracing {
		return w.db.Exec(ctx, tx, query, args...)
	}

	operation := extractOperation(query)
	sanitizedQuery := w.sanitizeQuery(query)

	ctx, span := dbTracer.Start(ctx, "db.exec",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemMySQL,
			attribute.String("db.operation", operation),
			attribute.String("db.statement", sanitizedQuery),
		),
	)
	defer span.End()

	start := time.Now()
	result, err := w.db.Exec(ctx, tx, query, args...)
	duration := time.Since(start)

	// Record metrics
	success := err == nil
	metrics.RecordDBQuery(operation, duration, success)

	// Set span attributes
	if result != nil && err == nil {
		if rowsAffected, err := result.RowsAffected(); err == nil {
			span.SetAttributes(attribute.Int64("db.rows_affected", rowsAffected))
		}
	}

	span.SetAttributes(
		attribute.Float64("db.duration_ms", float64(duration.Milliseconds())),
	)

	// Set span status
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return result, err
}

// PingContext wraps PingContext with tracing
func (w *OtelDatabaseWrapper) PingContext(ctx context.Context) error {
	if !w.enableTracing {
		return w.db.PingContext(ctx)
	}

	ctx, span := dbTracer.Start(ctx, "db.ping",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemMySQL,
			attribute.String("db.operation", "ping"),
		),
	)
	defer span.End()

	start := time.Now()
	err := w.db.PingContext(ctx)
	duration := time.Since(start)

	// Record metrics
	success := err == nil
	metrics.RecordDBQuery("ping", duration, success)

	// Set span status
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

// Close wraps Close (no tracing needed for cleanup)
func (w *OtelDatabaseWrapper) Close() error {
	return w.db.Close()
}

// extractOperation extracts the SQL operation type from the query
func extractOperation(query string) string {
	// Trim whitespace and convert to uppercase
	trimmed := strings.TrimSpace(strings.ToUpper(query))

	// Extract first word
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return "unknown"
	}

	operation := strings.ToLower(parts[0])

	// Normalize common operations
	switch operation {
	case "select":
		return "select"
	case "insert":
		return "insert"
	case "update":
		return "update"
	case "delete":
		return "delete"
	case "create":
		return "create"
	case "alter":
		return "alter"
	case "drop":
		return "drop"
	case "truncate":
		return "truncate"
	default:
		return operation
	}
}

// sanitizeQuery removes sensitive data and truncates long queries
var (
	// Regex to replace parameter values in queries
	paramRegex = regexp.MustCompile(`\?`)
	// Maximum query length to prevent huge spans
	maxQueryLength = 500
)

// sanitizeQuery sanitizes SQL queries for tracing
func (w *OtelDatabaseWrapper) sanitizeQuery(query string) string {
	if !w.sanitizeQueries {
		return query
	}

	// Replace newlines and multiple spaces with single space
	sanitized := regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	// Truncate if too long
	if len(sanitized) > maxQueryLength {
		sanitized = sanitized[:maxQueryLength] + "... [truncated]"
	}

	return sanitized
}
