package database

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const dbTracerName = "math-svr/db"

// startDBSpan opens a client span for one SQL statement as a child of the
// span in ctx (the request span). It returns a no-op span when tracing is
// disabled, so callers never branch.
//
// Safety: the recorded db.statement keeps its `?` placeholders — bound values
// live in the args slice and are NEVER attached to the span. This matters
// because the MySQL driver runs with InterpolateParams=true, so raw arg
// values must not leak into traces (see rules/security.md §3).
func startDBSpan(ctx context.Context, query string) trace.Span {
	op := sqlOp(query)
	_, span := otel.Tracer(dbTracerName).Start(ctx, "db.mysql "+op,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "mysql"),
			attribute.String("db.operation", op),
			attribute.String("db.statement", truncateStmt(query, 1024)),
		),
	)
	return span
}

// endDBSpan records the error (if any) and ends the span. Pass nil on success.
func endDBSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// sqlOp returns the leading SQL verb (SELECT/INSERT/UPDATE/DELETE/BEGIN/…)
// used as the span name suffix, uppercased.
func sqlOp(query string) string {
	q := strings.TrimSpace(query)
	if q == "" {
		return "UNKNOWN"
	}
	if i := strings.IndexAny(q, " \t\n"); i > 0 {
		return strings.ToUpper(q[:i])
	}
	return strings.ToUpper(q)
}

func truncateStmt(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
