package bot

import (
	"context"

	"math-ai.com/math-ai/internal/libs/openai"
)

// OpenAIProvider is the BotProvider implementation backed by libs/openai,
// which calls OpenAI's REST API directly through the project's shared
// http_client — no vendor SDK, no third party in the path. It owns nothing
// besides the wired *openai.Client and mirrors the other providers'
// translation duties.
//
// Why it exists alongside three providers that can already reach OpenAI:
// langchain and eino reach it through an SDK, and openrouter reaches it
// through a broker that rebills the call. This one is the only path where
// the request leaves the process addressed to api.openai.com with the
// project's own key — which is what makes usage, rate limits and the
// platform.openai.com/logs dashboard attributable to this account.
//
// Capability matrix: Chat, Stream and Embed are ALL supported. That makes
// it the only provider besides langchain that can serve embeddings.
//
// Note on Stream: libs/openai parses the SSE events out of a
// fully-buffered response, so chunk order and the assembled result are
// correct but delivery is not incremental. See
// openai.Client.GenerateStream.
type OpenAIProvider struct {
	client *openai.Client
}

// NewOpenAIProvider builds the provider from a constructed client.
func NewOpenAIProvider(client *openai.Client) *OpenAIProvider {
	return &OpenAIProvider{client: client}
}

func (p *OpenAIProvider) Name() BotProviderName { return ProviderOpenAI }

// Model exposes the configured default chat model id. Useful for log/audit
// surfaces that need to record which model served a request.
func (p *OpenAIProvider) Model() string { return p.client.Model() }

// Chat invokes openai.Client.Generate.
//
// The returned error is the typed MathError produced by mapOpenAIError;
// callers (adapter.Chat) propagate it unchanged.
func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	out, err := p.client.Generate(ctx, toOpenAIChat(req))
	if err != nil {
		return nil, mapOpenAIError(ctx, err)
	}
	return &ChatResult{
		Provider:     ProviderOpenAI,
		Model:        out.Model,
		Content:      out.Content,
		FinishReason: out.FinishReason,
		Usage:        toBotUsage(out.Usage),
	}, nil
}

// Stream invokes openai.Client.GenerateStream.
func (p *OpenAIProvider) Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error) {
	out, err := p.client.GenerateStream(ctx, toOpenAIChat(req), func(c openai.StreamChunk) error {
		var usage *Usage
		if c.Usage != nil {
			u := toBotUsage(*c.Usage)
			usage = &u
		}
		return onChunk(StreamChunk{
			Delta: c.Delta,
			Done:  c.Done,
			Usage: usage,
		})
	})
	if err != nil {
		return nil, mapOpenAIError(ctx, err)
	}
	return &ChatResult{
		Provider:     ProviderOpenAI,
		Model:        out.Model,
		Content:      out.Content,
		FinishReason: out.FinishReason,
		Usage:        toBotUsage(out.Usage),
	}, nil
}

// Embed invokes openai.Client.Embed against /v1/embeddings. Unlike the
// eino and openrouter providers, this is a real implementation rather than
// a BOT_UNSUPPORTED_OP stub.
func (p *OpenAIProvider) Embed(ctx context.Context, req EmbedRequest) (*EmbedResult, error) {
	out, err := p.client.Embed(ctx, openai.EmbedRequest{
		Model:  req.Model,
		Inputs: req.Inputs,
	})
	if err != nil {
		return nil, mapOpenAIError(ctx, err)
	}
	return &EmbedResult{
		Provider: ProviderOpenAI,
		Model:    out.Model,
		Vectors:  out.Vectors,
		Usage:    toBotUsage(out.Usage),
	}, nil
}

func toBotUsage(u openai.Usage) Usage {
	return Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}
}

func toOpenAIChat(req ChatRequest) openai.ChatRequest {
	msgs := make([]openai.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, openai.Message{
			Role:    toOpenAIRole(m.Role),
			Content: m.Content,
			Name:    m.Name,
		})
	}
	return openai.ChatRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stop:        req.Stop,
		JSONMode:    req.JSONMode,
	}
}

func toOpenAIRole(r Role) openai.Role {
	switch r {
	case RoleSystem:
		return openai.RoleSystem
	case RoleAssistant:
		return openai.RoleAssistant
	case RoleTool:
		return openai.RoleTool
	default:
		return openai.RoleUser
	}
}
