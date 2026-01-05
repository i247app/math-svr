package ai_provider

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
	"math-ai.com/math-ai/internal/shared/telemetry/metrics"
)

var aiTracer = otel.Tracer("math-srv/ai-provider")

// OtelAIClientWrapper wraps IChatBoxClient with OpenTelemetry instrumentation
type OtelAIClientWrapper struct {
	client         IChatBoxClient
	providerName   string
	enableTracing  bool
}

// NewOtelAIClientWrapper creates a new instrumented AI client wrapper
func NewOtelAIClientWrapper(client IChatBoxClient, providerName string, enableTracing bool) *OtelAIClientWrapper {
	return &OtelAIClientWrapper{
		client:        client,
		providerName:  providerName,
		enableTracing: enableTracing,
	}
}

// SendMessage wraps SendMessage with tracing and metrics
func (w *OtelAIClientWrapper) SendMessage(ctx context.Context, conv *domain.Conversation) (*ChatCompletionResponse, error) {
	if !w.enableTracing {
		return w.client.SendMessage(ctx, conv)
	}

	// Create span
	ctx, span := aiTracer.Start(ctx, "ai.send_message",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("ai.provider", w.providerName),
			attribute.String("ai.operation", "send_message"),
			attribute.String("ai.model", conv.Model()),
			attribute.Float64("ai.temperature", float64(conv.Temperature())),
			attribute.Int("ai.max_tokens", conv.MaxTokens()),
			attribute.Int("ai.message_count", len(conv.Messages())),
		),
	)
	defer span.End()

	// Add system prompt info if present
	if conv.SystemPrompt() != nil && *conv.SystemPrompt() != "" {
		span.SetAttributes(attribute.Bool("ai.has_system_prompt", true))
	}

	start := time.Now()
	response, err := w.client.SendMessage(ctx, conv)
	duration := time.Since(start)

	// Record metrics
	success := err == nil
	var promptTokens, completionTokens int
	if response != nil {
		promptTokens = response.PromptTokens
		completionTokens = response.CompletionTokens
	}
	metrics.RecordAIRequest(w.providerName, "send_message", conv.Model(), duration, success, promptTokens, completionTokens)

	// Set span attributes from response
	if response != nil {
		span.SetAttributes(
			attribute.Int("ai.prompt_tokens", response.PromptTokens),
			attribute.Int("ai.completion_tokens", response.CompletionTokens),
			attribute.Int("ai.total_tokens", response.TotalTokens),
			attribute.String("ai.finish_reason", response.FinishReason),
			attribute.Int("ai.response_length", len(response.Message)),
		)
	}

	span.SetAttributes(
		attribute.Float64("ai.duration_ms", float64(duration.Milliseconds())),
	)

	// Set span status
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return response, err
}

// StreamMessage wraps StreamMessage with tracing and metrics
func (w *OtelAIClientWrapper) StreamMessage(ctx context.Context, conv *domain.Conversation) (<-chan StreamChunk, error) {
	if !w.enableTracing {
		return w.client.StreamMessage(ctx, conv)
	}

	// Create span for the stream initialization
	ctx, span := aiTracer.Start(ctx, "ai.stream_message",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("ai.provider", w.providerName),
			attribute.String("ai.operation", "stream_message"),
			attribute.String("ai.model", conv.Model()),
			attribute.Float64("ai.temperature", float64(conv.Temperature())),
			attribute.Int("ai.max_tokens", conv.MaxTokens()),
			attribute.Int("ai.message_count", len(conv.Messages())),
		),
	)
	defer span.End()

	// Add system prompt info if present
	if conv.SystemPrompt() != nil && *conv.SystemPrompt() != "" {
		span.SetAttributes(attribute.Bool("ai.has_system_prompt", true))
	}

	start := time.Now()
	chunkChan, err := w.client.StreamMessage(ctx, conv)

	if err != nil {
		duration := time.Since(start)
		// Record failed stream initialization
		metrics.RecordAIRequest(w.providerName, "stream_message", conv.Model(), duration, false, 0, 0)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Create a wrapper channel to track streaming metrics
	wrappedChan := make(chan StreamChunk)

	go func() {
		defer close(wrappedChan)

		var totalChunks int
		var totalLength int
		var streamErr error
		var finishReason string

		for chunk := range chunkChan {
			// Forward the chunk
			wrappedChan <- chunk

			// Track metrics
			if chunk.Error != nil {
				streamErr = chunk.Error
			}
			if chunk.Done {
				finishReason = chunk.FinishReason
				break
			}
			totalChunks++
			totalLength += len(chunk.Delta)
		}

		duration := time.Since(start)

		// Record final metrics
		success := streamErr == nil
		// For streaming, we don't have exact token counts, so estimate
		estimatedTokens := totalLength / 4 // Rough estimation
		metrics.RecordAIRequest(w.providerName, "stream_message", conv.Model(), duration, success, 0, estimatedTokens)

		// Update span with streaming statistics
		span.SetAttributes(
			attribute.Int("ai.stream.chunk_count", totalChunks),
			attribute.Int("ai.stream.total_length", totalLength),
			attribute.String("ai.finish_reason", finishReason),
			attribute.Float64("ai.duration_ms", float64(duration.Milliseconds())),
		)

		if streamErr != nil {
			span.RecordError(streamErr)
			span.SetStatus(codes.Error, streamErr.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}()

	return wrappedChan, nil
}
