package bot

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/deepseek"
)

// DeepSeekProvider is the BotProvider implementation backed by
// libs/deepseek, which calls api.deepseek.com directly through the
// project's shared http_client — no vendor SDK, no third party in the
// path. It owns nothing besides the wired *deepseek.Client.
//
// Why it exists alongside providers that can already reach DeepSeek:
// openrouter reaches it through a broker that rebills the call, and the
// SDK-backed providers reach it only by pointing an OpenAI-compatible
// backend at DeepSeek's base URL. This is the only path where the request
// leaves the process addressed to DeepSeek with the project's own key, so
// balance, rate limits and billing are attributable to this account.
//
// Capability matrix: Chat and Stream are supported. Embed is NOT — the
// platform has no embeddings endpoint, so it returns
// MathError(BOT_UNSUPPORTED_OP) per the BotProvider contract for optional
// capabilities, the same posture as the eino and openrouter providers.
//
// Note on Stream: libs/deepseek parses the SSE events out of a
// fully-buffered response, so chunk order and the assembled result are
// correct but delivery is not incremental. See
// deepseek.Client.GenerateStream.
type DeepSeekProvider struct {
	client *deepseek.Client
}

// NewDeepSeekProvider builds the provider from a constructed client.
func NewDeepSeekProvider(client *deepseek.Client) *DeepSeekProvider {
	return &DeepSeekProvider{client: client}
}

func (p *DeepSeekProvider) Name() BotProviderName { return ProviderDeepSeek }

// Model exposes the configured default model id. Useful for log/audit
// surfaces that need to record which model served a request.
func (p *DeepSeekProvider) Model() string { return p.client.Model() }

// Chat invokes deepseek.Client.Generate.
//
// The returned error is the typed MathError produced by mapDeepSeekError;
// callers (adapter.Chat) propagate it unchanged.
//
// ReasoningContent is deliberately dropped here: the adapter's ChatResult
// has no field for it, and callers in this repo want the answer (a quiz
// JSON blob), not the model's deliberation. The reasoning TOKEN COUNT is
// still logged by libs/deepseek, so the cost stays visible.
func (p *DeepSeekProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	out, err := p.client.Generate(ctx, toDeepSeekChat(req))
	if err != nil {
		return nil, mapDeepSeekError(ctx, err)
	}
	return &ChatResult{
		Provider:     ProviderDeepSeek,
		Model:        out.Model,
		Content:      out.Content,
		FinishReason: out.FinishReason,
		Usage:        fromDeepSeekUsage(out.Usage),
	}, nil
}

// Stream invokes deepseek.Client.GenerateStream.
func (p *DeepSeekProvider) Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error) {
	out, err := p.client.GenerateStream(ctx, toDeepSeekChat(req), func(c deepseek.StreamChunk) error {
		var usage *Usage
		if c.Usage != nil {
			u := fromDeepSeekUsage(*c.Usage)
			usage = &u
		}
		return onChunk(StreamChunk{
			Delta: c.Delta,
			Done:  c.Done,
			Usage: usage,
		})
	})
	if err != nil {
		return nil, mapDeepSeekError(ctx, err)
	}
	return &ChatResult{
		Provider:     ProviderDeepSeek,
		Model:        out.Model,
		Content:      out.Content,
		FinishReason: out.FinishReason,
		Usage:        fromDeepSeekUsage(out.Usage),
	}, nil
}

// Embed reports the capability as unsupported. DeepSeek exposes no
// embedding endpoint; deploys that need embeddings should route Embed via
// the openai, gemini or langchain providers (EmbedVia).
func (p *DeepSeekProvider) Embed(ctx context.Context, _ EmbedRequest) (*EmbedResult, error) {
	return nil, errs.NewError(ctx, status.BOT_UNSUPPORTED_OP,
		map[string]any{"provider": string(ProviderDeepSeek)},
		errors.New("deepseek provider does not support embeddings"))
}

// fromDeepSeekUsage narrows to the adapter's shared Usage. The cache-hit,
// cache-miss and reasoning counts have no field on that type; they are
// logged inside libs/deepseek instead of being propagated, because they
// explain a bill rather than change a caller's behaviour.
func fromDeepSeekUsage(u deepseek.Usage) Usage {
	return Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}
}

func toDeepSeekChat(req ChatRequest) deepseek.ChatRequest {
	msgs := make([]deepseek.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, deepseek.Message{
			Role:    toDeepSeekRole(m.Role),
			Content: m.Content,
			Name:    m.Name,
		})
	}
	return deepseek.ChatRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stop:        req.Stop,
		JSONMode:    req.JSONMode,
	}
}

func toDeepSeekRole(r Role) deepseek.Role {
	switch r {
	case RoleSystem:
		return deepseek.RoleSystem
	case RoleAssistant:
		return deepseek.RoleAssistant
	case RoleTool:
		return deepseek.RoleTool
	default:
		return deepseek.RoleUser
	}
}
