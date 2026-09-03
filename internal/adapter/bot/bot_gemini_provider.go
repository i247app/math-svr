package bot

import (
	"context"

	"math-ai.com/math-ai/internal/libs/gemini"
)

// GeminiProvider is the BotProvider implementation backed by libs/gemini,
// which calls generativelanguage.googleapis.com directly through the
// project's shared http_client — no vendor SDK, no third party in the
// path. It owns nothing besides the wired *gemini.Client.
//
// Why it exists alongside providers that can already reach Gemini:
// langchain and eino reach it through an SDK (googleai backend), and
// openrouter reaches it through a broker that rebills the call. This is
// the only path where the request leaves the process addressed to Google
// with the project's own key, so quota, rate limits and billing are
// attributable to this account.
//
// Capability matrix: Chat, Stream and Embed are ALL supported.
//
// Note on Stream: libs/gemini parses the SSE events out of a
// fully-buffered response, so chunk order and the assembled result are
// correct but delivery is not incremental. See
// gemini.Client.GenerateStream.
type GeminiProvider struct {
	client *gemini.Client
}

// NewGeminiProvider builds the provider from a constructed client.
func NewGeminiProvider(client *gemini.Client) *GeminiProvider {
	return &GeminiProvider{client: client}
}

func (p *GeminiProvider) Name() BotProviderName { return ProviderGemini }

// Model exposes the configured default model id. Useful for log/audit
// surfaces that need to record which model served a request.
func (p *GeminiProvider) Model() string { return p.client.Model() }

// Chat invokes gemini.Client.Generate.
//
// The returned error is the typed MathError produced by mapGeminiError;
// callers (adapter.Chat) propagate it unchanged.
func (p *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	out, err := p.client.Generate(ctx, toGeminiChat(req))
	if err != nil {
		return nil, mapGeminiError(ctx, err)
	}
	return &ChatResult{
		Provider:     ProviderGemini,
		Model:        out.Model,
		Content:      out.Content,
		FinishReason: out.FinishReason,
		Usage:        fromGeminiUsage(out.Usage),
	}, nil
}

// Stream invokes gemini.Client.GenerateStream.
func (p *GeminiProvider) Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error) {
	out, err := p.client.GenerateStream(ctx, toGeminiChat(req), func(c gemini.StreamChunk) error {
		var usage *Usage
		if c.Usage != nil {
			u := fromGeminiUsage(*c.Usage)
			usage = &u
		}
		return onChunk(StreamChunk{
			Delta: c.Delta,
			Done:  c.Done,
			Usage: usage,
		})
	})
	if err != nil {
		return nil, mapGeminiError(ctx, err)
	}
	return &ChatResult{
		Provider:     ProviderGemini,
		Model:        out.Model,
		Content:      out.Content,
		FinishReason: out.FinishReason,
		Usage:        fromGeminiUsage(out.Usage),
	}, nil
}

// Embed invokes gemini.Client.Embed against :batchEmbedContents.
func (p *GeminiProvider) Embed(ctx context.Context, req EmbedRequest) (*EmbedResult, error) {
	out, err := p.client.Embed(ctx, gemini.EmbedRequest{
		Model:  req.Model,
		Inputs: req.Inputs,
	})
	if err != nil {
		return nil, mapGeminiError(ctx, err)
	}
	return &EmbedResult{
		Provider: ProviderGemini,
		Model:    out.Model,
		Vectors:  out.Vectors,
		Usage:    fromGeminiUsage(out.Usage),
	}, nil
}

func fromGeminiUsage(u gemini.Usage) Usage {
	return Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}
}

func toGeminiChat(req ChatRequest) gemini.ChatRequest {
	msgs := make([]gemini.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, gemini.Message{
			Role:    toGeminiRole(m.Role),
			Content: m.Content,
			Name:    m.Name,
		})
	}
	return gemini.ChatRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stop:        req.Stop,
		JSONMode:    req.JSONMode,
	}
}

// toGeminiRole keeps the adapter's four roles intact across the boundary.
// The narrowing onto Gemini's user/model pair — and lifting system out
// into systemInstruction — happens inside gemini.Client.buildRequest,
// where the whole message list is visible at once.
func toGeminiRole(r Role) gemini.Role {
	switch r {
	case RoleSystem:
		return gemini.RoleSystem
	case RoleAssistant:
		return gemini.RoleAssistant
	case RoleTool:
		return gemini.RoleTool
	default:
		return gemini.RoleUser
	}
}
