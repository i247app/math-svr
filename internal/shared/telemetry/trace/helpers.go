package trace

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var serviceTracer = otel.Tracer("math-srv/service")

// StartServiceSpan starts a new span for a service operation
// Use this at the beginning of important service methods to create custom spans
//
// Example:
//
//	func (s *UserService) ProcessPayment(ctx context.Context, userID string, amount float64) error {
//	    ctx, span := trace.StartServiceSpan(ctx, "UserService.ProcessPayment",
//	        attribute.String("user_id", userID),
//	        attribute.Float64("amount", amount),
//	    )
//	    defer span.End()
//
//	    // Your service logic here...
//	    if err != nil {
//	        trace.RecordServiceError(span, err)
//	        return err
//	    }
//
//	    trace.MarkServiceSuccess(span)
//	    return nil
//	}
func StartServiceSpan(ctx context.Context, operationName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := serviceTracer.Start(ctx, operationName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attrs...),
	)
	return ctx, span
}

// StartServiceSpanWithParent starts a new span with an explicit parent span
// Useful when you need to manually control span hierarchy
func StartServiceSpanWithParent(ctx context.Context, parentSpan trace.Span, operationName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx = trace.ContextWithSpan(ctx, parentSpan)
	return StartServiceSpan(ctx, operationName, attrs...)
}

// RecordServiceError records an error in a span and sets the error status
// Use this when an operation fails
func RecordServiceError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	span.SetAttributes(attribute.Bool("service.error", true))
}

// RecordServiceErrorWithMessage records an error with a custom message
func RecordServiceErrorWithMessage(span trace.Span, err error, message string) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, message)
	span.SetAttributes(
		attribute.Bool("service.error", true),
		attribute.String("service.error_message", message),
	)
}

// MarkServiceSuccess marks a span as successfully completed
// Use this when an operation completes successfully
func MarkServiceSuccess(span trace.Span) {
	span.SetStatus(codes.Ok, "")
	span.SetAttributes(attribute.Bool("service.success", true))
}

// AddServiceAttribute adds a single attribute to a span
func AddServiceAttribute(span trace.Span, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	default:
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}

// AddServiceAttributes adds multiple attributes to a span
func AddServiceAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// RecordServiceEvent records a discrete event in a span's timeline
// Useful for marking important moments in a long-running operation
//
// Example:
//
//	trace.RecordServiceEvent(span, "validation.completed")
//	trace.RecordServiceEvent(span, "database.query.started")
//	trace.RecordServiceEvent(span, "external.api.called", attribute.String("api", "stripe"))
func RecordServiceEvent(span trace.Span, eventName string, attrs ...attribute.KeyValue) {
	span.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// StartAsyncOperation creates a span for an asynchronous operation
// Returns a new context that should be used in the goroutine
//
// Example:
//
//	func (s *Service) ProcessAsync(ctx context.Context, data string) {
//	    go func() {
//	        asyncCtx, span := trace.StartAsyncOperation(ctx, "Service.ProcessAsync")
//	        defer span.End()
//
//	        // Async work here...
//	    }()
//	}
func StartAsyncOperation(ctx context.Context, operationName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	// Extract the span context but don't make it the active span
	// This prevents the async operation from affecting the parent's span
	spanContext := trace.SpanContextFromContext(ctx)
	newCtx := trace.ContextWithSpanContext(context.Background(), spanContext)

	return StartServiceSpan(newCtx, operationName, attrs...)
}

// WithServiceSpan is a helper that wraps a function with a span
// Automatically handles span creation, error recording, and cleanup
//
// Example:
//
//	err := trace.WithServiceSpan(ctx, "calculateTotal", func(ctx context.Context, span trace.Span) error {
//	    total := calculateSomething()
//	    span.SetAttributes(attribute.Float64("total", total))
//	    return nil
//	})
func WithServiceSpan(ctx context.Context, operationName string, fn func(context.Context, trace.Span) error, attrs ...attribute.KeyValue) error {
	ctx, span := StartServiceSpan(ctx, operationName, attrs...)
	defer span.End()

	err := fn(ctx, span)
	if err != nil {
		RecordServiceError(span, err)
		return err
	}

	MarkServiceSuccess(span)
	return nil
}

// GetSpanFromContext retrieves the current span from context
// Returns a non-nil span even if no span exists (no-op span)
func GetSpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// IsSpanRecording checks if the current span is recording
// Useful to avoid expensive operations when tracing is disabled
func IsSpanRecording(ctx context.Context) bool {
	return trace.SpanFromContext(ctx).IsRecording()
}
