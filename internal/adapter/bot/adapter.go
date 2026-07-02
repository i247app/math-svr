package bot

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const botTracerName = "math-svr/bot"

// Adapter is the entry point callers use to dispatch AI requests. It
// holds one or more registered providers and a designated default used
// by Chat / Stream / Embed.
//
// The Adapter is safe for concurrent use after construction; provider
// registration is expected to happen at boot inside NewFromConfig.
type Adapter struct {
	providers   map[BotProviderName]BotProvider
	defaultName BotProviderName
}

func NewAdapter() *Adapter {
	return &Adapter{providers: make(map[BotProviderName]BotProvider)}
}

// Register adds a provider to the adapter. The first provider registered
// becomes the default; callers may override that with SetDefault.
func (a *Adapter) Register(provider BotProvider) {
	if a.providers == nil {
		a.providers = make(map[BotProviderName]BotProvider)
	}
	a.providers[provider.Name()] = provider
	if a.defaultName == "" {
		a.defaultName = provider.Name()
	}
}

// SetDefault selects which registered provider Chat / Stream / Embed
// dispatch to when no explicit name is supplied. Returns an error if
// the provider name has not been registered.
func (a *Adapter) SetDefault(name BotProviderName) error {
	if _, ok := a.providers[name]; !ok {
		return fmt.Errorf("bot: provider %q is not registered", name)
	}
	a.defaultName = name
	return nil
}

// DefaultName returns the currently designated default provider name.
// Empty when no providers are registered.
func (a *Adapter) DefaultName() BotProviderName { return a.defaultName }

// Has reports whether name is registered.
func (a *Adapter) Has(name BotProviderName) bool {
	_, ok := a.providers[name]
	return ok
}

// Chat dispatches req through the default provider after running
// ChatRequest.Validate.
func (a *Adapter) Chat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	if a.defaultName == "" {
		return nil, errors.New("bot: no default provider configured")
	}
	return a.ChatVia(ctx, a.defaultName, req)
}

// ChatVia dispatches req through the provider registered under name.
//
// Validation happens here so failures are caught uniformly regardless
// of routing entry point; the underlying provider is never called on
// an invalid payload, which keeps vendor billing clean.
func (a *Adapter) ChatVia(ctx context.Context, name BotProviderName, req ChatRequest) (*ChatResult, error) {
	log := logger.From(ctx)

	if err := req.Validate(); err != nil {
		return nil, errs.NewError(ctx, status.BOT_INVALID_PROMPT,
			map[string]any{"reason": err.Error()}, err)
	}

	provider, ok := a.providers[name]
	if !ok {
		return nil, fmt.Errorf("bot: provider %q is not registered", name)
	}

	ctx, span := startLLMSpan(ctx, name, "chat")
	defer span.End()

	res, err := provider.Chat(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, logAndReturn(ctx, log, name, "chat", err)
	}
	setLLMResultAttrs(span, res.Model, res.Usage.PromptTokens, res.Usage.CompletionTokens, res.Usage.TotalTokens)
	log.Infof("bot.chat provider=%s model=%s prompt_tokens=%d completion_tokens=%d total_tokens=%d",
		name, res.Model, res.Usage.PromptTokens, res.Usage.CompletionTokens, res.Usage.TotalTokens)
	return res, nil
}

// Stream dispatches a streaming chat request through the default
// provider after running ChatRequest.Validate. onChunk MUST NOT block —
// the caller's pipeline owns downstream buffering.
func (a *Adapter) Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error) {
	if a.defaultName == "" {
		return nil, errors.New("bot: no default provider configured")
	}
	return a.StreamVia(ctx, a.defaultName, req, onChunk)
}

// StreamVia dispatches a streaming chat request through the provider
// registered under name.
func (a *Adapter) StreamVia(ctx context.Context, name BotProviderName, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error) {
	log := logger.From(ctx)

	if onChunk == nil {
		return nil, errs.NewError(ctx, status.BOT_INVALID_PROMPT, nil,
			errors.New("bot: onChunk is required for streaming"))
	}
	if err := req.Validate(); err != nil {
		return nil, errs.NewError(ctx, status.BOT_INVALID_PROMPT,
			map[string]any{"reason": err.Error()}, err)
	}

	provider, ok := a.providers[name]
	if !ok {
		return nil, fmt.Errorf("bot: provider %q is not registered", name)
	}

	ctx, span := startLLMSpan(ctx, name, "stream")
	defer span.End()

	res, err := provider.Stream(ctx, req, onChunk)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, logAndReturn(ctx, log, name, "stream", err)
	}
	setLLMResultAttrs(span, res.Model, res.Usage.PromptTokens, res.Usage.CompletionTokens, res.Usage.TotalTokens)
	log.Infof("bot.stream provider=%s model=%s prompt_tokens=%d completion_tokens=%d",
		name, res.Model, res.Usage.PromptTokens, res.Usage.CompletionTokens)
	return res, nil
}

// Embed dispatches req through the default provider after running
// EmbedRequest.Validate.
func (a *Adapter) Embed(ctx context.Context, req EmbedRequest) (*EmbedResult, error) {
	if a.defaultName == "" {
		return nil, errors.New("bot: no default provider configured")
	}
	return a.EmbedVia(ctx, a.defaultName, req)
}

// EmbedVia dispatches req through the provider registered under name.
func (a *Adapter) EmbedVia(ctx context.Context, name BotProviderName, req EmbedRequest) (*EmbedResult, error) {
	log := logger.From(ctx)

	if err := req.Validate(); err != nil {
		return nil, errs.NewError(ctx, status.BOT_INVALID_PROMPT,
			map[string]any{"reason": err.Error()}, err)
	}

	provider, ok := a.providers[name]
	if !ok {
		return nil, fmt.Errorf("bot: provider %q is not registered", name)
	}

	ctx, span := startLLMSpan(ctx, name, "embed")
	defer span.End()

	res, err := provider.Embed(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, logAndReturn(ctx, log, name, "embed", err)
	}
	span.SetAttributes(attribute.String("gen_ai.response.model", res.Model))
	log.Infof("bot.embed provider=%s model=%s inputs=%d", name, res.Model, len(req.Inputs))
	return res, nil
}

// startLLMSpan opens a client span for one LLM operation as a child of the
// span in ctx. Returns a no-op span when tracing is disabled.
//
// Log/trace discipline: only operator metadata (provider, operation, model,
// token counts) is attached — never the prompt content or the response body.
func startLLMSpan(ctx context.Context, name BotProviderName, op string) (context.Context, trace.Span) {
	return otel.Tracer(botTracerName).Start(ctx, "llm."+op,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gen_ai.system", string(name)),
			attribute.String("gen_ai.operation.name", op),
		),
	)
}

// setLLMResultAttrs stamps the model + token usage onto a finished span.
func setLLMResultAttrs(span trace.Span, model string, promptTokens, completionTokens, totalTokens int) {
	span.SetAttributes(
		attribute.String("gen_ai.response.model", model),
		attribute.Int("gen_ai.usage.input_tokens", promptTokens),
		attribute.Int("gen_ai.usage.output_tokens", completionTokens),
		attribute.Int("gen_ai.usage.total_tokens", totalTokens),
	)
}

// logAndReturn is the common error-finalisation path: classify, log at
// the right severity, and return the MathError unchanged for the caller.
//
// Log discipline: never log the prompt content, never the upstream
// response body, never any credential material. Operator context only.
func logAndReturn(ctx context.Context, log interface {
	Errorf(string, ...any)
	Warnf(string, ...any)
}, name BotProviderName, op string, err error) error {
	mErr, ok := errs.IsMathError(err)
	if !ok {
		// A non-MathError leaking up here means a provider bypassed the
		// translation contract. Surface as BOT_OP_FAILED so the wire
		// shape stays uniform; log loudly to flag the regression.
		log.Errorf("bot.%s_error provider=%s err=%v (provider returned non-MathError)", op, name, err)
		return errs.NewError(ctx, status.BOT_OP_FAILED,
			map[string]any{"provider": string(name)}, err)
	}

	switch mErr.GetStatusCode() {
	case status.BOT_CONFIG_INVALID, status.BOT_UNSUPPORTED_OP:
		log.Errorf("bot.%s_config_error provider=%s err=%v", op, name, err)
	case status.BOT_SERIALIZE_FAILED:
		log.Errorf("bot.%s_serialize_error provider=%s err=%v", op, name, err)
	case status.BOT_RATE_LIMITED:
		log.Warnf("bot.%s_rate_limited provider=%s err=%v", op, name, err)
	case status.BOT_CONTEXT_TOO_LARGE:
		log.Warnf("bot.%s_context_too_large provider=%s err=%v", op, name, err)
	default:
		// BOT_OP_FAILED / BOT_CONNECT_FAILED — transient or upstream-noise.
		log.Warnf("bot.%s_op_error provider=%s err=%v", op, name, err)
	}
	return mErr
}
