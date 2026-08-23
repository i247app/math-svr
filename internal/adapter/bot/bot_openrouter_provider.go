package bot

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/openrouter"
)

// OpenRouterProvider is the BotProvider implementation backed by
// libs/openrouter, which calls the OpenRouter REST API directly through
// the project's shared http_client — no vendor SDK involved. It owns
// nothing besides the wired *openrouter.Client and mirrors
// EinoProvider's translation duties.
//
// The single OpenRouterProvider can serve any model on OpenRouter; the
// vendor is chosen by the model id ("vendor/model-name") configured at
// boot and exposed via Model().
//
// Capability matrix: Chat and Stream are supported. Embed is NOT — it
// returns MathError(BOT_UNSUPPORTED_OP) per the BotProvider contract for
// optional capabilities (OpenRouter's chat-completions API has no
// embedding endpoint).
//
// Note on Stream: libs/openrouter parses the SSE events out of a
// fully-buffered response, so chunk ORDER and the assembled result are
// correct but delivery is not incremental. See
// openrouter.Client.GenerateStream.
type OpenRouterProvider struct {
	client *openrouter.Client
}

// NewOpenRouterProvider builds the provider from a constructed client.
func NewOpenRouterProvider(client *openrouter.Client) *OpenRouterProvider {
	return &OpenRouterProvider{client: client}
}

func (p *OpenRouterProvider) Name() BotProviderName { return ProviderOpenRouter }

// Model exposes the configured default model id. Useful for log/audit
// surfaces that need to record which model served a request.
func (p *OpenRouterProvider) Model() string { return p.client.Model() }

// Chat invokes openrouter.Client.Generate.
//
// The returned error is the typed MathError produced by
// mapOpenRouterError; callers (adapter.Chat) propagate it unchanged.
func (p *OpenRouterProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	out, err := p.client.Generate(ctx, toOpenRouterChat(req))
	if err != nil {
		return nil, mapOpenRouterError(ctx, err)
	}
	return &ChatResult{
		Provider:     ProviderOpenRouter,
		Model:        out.Model,
		Content:      out.Content,
		FinishReason: out.FinishReason,
		Usage: Usage{
			PromptTokens:     out.Usage.PromptTokens,
			CompletionTokens: out.Usage.CompletionTokens,
			TotalTokens:      out.Usage.TotalTokens,
		},
	}, nil
}

// Stream invokes openrouter.Client.GenerateStream.
func (p *OpenRouterProvider) Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error) {
	out, err := p.client.GenerateStream(ctx, toOpenRouterChat(req), func(c openrouter.StreamChunk) error {
		var usage *Usage
		if c.Usage != nil {
			usage = &Usage{
				PromptTokens:     c.Usage.PromptTokens,
				CompletionTokens: c.Usage.CompletionTokens,
				TotalTokens:      c.Usage.TotalTokens,
			}
		}
		return onChunk(StreamChunk{
			Delta: c.Delta,
			Done:  c.Done,
			Usage: usage,
		})
	})
	if err != nil {
		return nil, mapOpenRouterError(ctx, err)
	}
	return &ChatResult{
		Provider:     ProviderOpenRouter,
		Model:        out.Model,
		Content:      out.Content,
		FinishReason: out.FinishReason,
		Usage: Usage{
			PromptTokens:     out.Usage.PromptTokens,
			CompletionTokens: out.Usage.CompletionTokens,
			TotalTokens:      out.Usage.TotalTokens,
		},
	}, nil
}

// Embed reports the capability as unsupported. OpenRouter exposes no
// embedding endpoint; deploys that need embeddings should route Embed via
// the langchain provider (EmbedVia).
func (p *OpenRouterProvider) Embed(ctx context.Context, _ EmbedRequest) (*EmbedResult, error) {
	return nil, errs.NewError(ctx, status.BOT_UNSUPPORTED_OP,
		map[string]any{"provider": string(ProviderOpenRouter)},
		errors.New("openrouter provider does not support embeddings"))
}

func toOpenRouterChat(req ChatRequest) openrouter.ChatRequest {
	msgs := make([]openrouter.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, openrouter.Message{
			Role:    toOpenRouterRole(m.Role),
			Content: m.Content,
			Name:    m.Name,
		})
	}
	return openrouter.ChatRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stop:        req.Stop,
		JSONMode:    req.JSONMode,
	}
}

func toOpenRouterRole(r Role) openrouter.Role {
	switch r {
	case RoleSystem:
		return openrouter.RoleSystem
	case RoleAssistant:
		return openrouter.RoleAssistant
	case RoleTool:
		return openrouter.RoleTool
	default:
		return openrouter.RoleUser
	}
}
