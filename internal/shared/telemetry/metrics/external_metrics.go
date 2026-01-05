package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	// AI provider metrics
	aiRequestCounter metric.Int64Counter
	aiRequestDuration metric.Float64Histogram
	aiTokenCounter metric.Int64Counter

	// S3 metrics
	s3OperationCounter metric.Int64Counter
	s3OperationDuration metric.Float64Histogram

	// SMS metrics
	smsCounter metric.Int64Counter

	// Email metrics
	emailCounter metric.Int64Counter
)

func init() {
	var err error

	// AI metrics
	aiRequestCounter, err = meter.Int64Counter(
		"ai.request.count",
		metric.WithDescription("Total number of AI provider requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		panic(err)
	}

	aiRequestDuration, err = meter.Float64Histogram(
		"ai.request.duration",
		metric.WithDescription("AI provider request duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic(err)
	}

	aiTokenCounter, err = meter.Int64Counter(
		"ai.token.usage",
		metric.WithDescription("Total AI tokens used (prompt and completion)"),
		metric.WithUnit("{token}"),
	)
	if err != nil {
		panic(err)
	}

	// S3 metrics
	s3OperationCounter, err = meter.Int64Counter(
		"s3.operation.count",
		metric.WithDescription("Total number of S3 operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		panic(err)
	}

	s3OperationDuration, err = meter.Float64Histogram(
		"s3.operation.duration",
		metric.WithDescription("S3 operation duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic(err)
	}

	// SMS metrics
	smsCounter, err = meter.Int64Counter(
		"sms.sent.count",
		metric.WithDescription("Total number of SMS messages sent"),
		metric.WithUnit("{message}"),
	)
	if err != nil {
		panic(err)
	}

	// Email metrics
	emailCounter, err = meter.Int64Counter(
		"email.sent.count",
		metric.WithDescription("Total number of emails sent"),
		metric.WithUnit("{email}"),
	)
	if err != nil {
		panic(err)
	}
}

// RecordAIRequest records metrics for an AI provider request
func RecordAIRequest(provider, operation, model string, duration time.Duration, success bool, promptTokens, completionTokens int) {
	ctx := context.Background()

	status := "success"
	if !success {
		status = "error"
	}

	attrs := []attribute.KeyValue{
		attribute.String("ai.provider", provider),
		attribute.String("ai.operation", operation),
		attribute.String("ai.model", model),
		attribute.String("ai.status", status),
	}

	// Record request count
	aiRequestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Record request duration
	aiRequestDuration.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))

	// Record token usage
	if promptTokens > 0 {
		aiTokenCounter.Add(ctx, int64(promptTokens), metric.WithAttributes(
			attribute.String("ai.provider", provider),
			attribute.String("ai.token_type", "prompt"),
		))
	}
	if completionTokens > 0 {
		aiTokenCounter.Add(ctx, int64(completionTokens), metric.WithAttributes(
			attribute.String("ai.provider", provider),
			attribute.String("ai.token_type", "completion"),
		))
	}
}

// RecordS3Operation records metrics for an S3 operation
func RecordS3Operation(operation string, duration time.Duration, success bool) {
	ctx := context.Background()

	status := "success"
	if !success {
		status = "error"
	}

	attrs := []attribute.KeyValue{
		attribute.String("s3.operation", operation),
		attribute.String("s3.status", status),
	}

	s3OperationCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	s3OperationDuration.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))
}

// RecordSMS records metrics for SMS sending
func RecordSMS(success bool, recipientCount int) {
	ctx := context.Background()

	status := "success"
	if !success {
		status = "error"
	}

	attrs := []attribute.KeyValue{
		attribute.String("sms.status", status),
	}

	smsCounter.Add(ctx, int64(recipientCount), metric.WithAttributes(attrs...))
}

// RecordEmail records metrics for email sending
func RecordEmail(success bool, recipientCount int) {
	ctx := context.Background()

	status := "success"
	if !success {
		status = "error"
	}

	attrs := []attribute.KeyValue{
		attribute.String("email.status", status),
	}

	emailCounter.Add(ctx, int64(recipientCount), metric.WithAttributes(attrs...))
}
