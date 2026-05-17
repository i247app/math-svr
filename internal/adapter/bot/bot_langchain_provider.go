package bot

import (
	"context"

	"math-ai.com/math-ai/internal/libs/langchain"
)

// LangChainProvider is the BotProvider implementation backed by
// libs/langchain. It owns nothing besides the wired *langchain.Client.
//
// The single LangChainProvider can serve any of the langchaingo-backed
// LLM vendors (googleai, openai, anthropic, ollama) — the choice is
// fixed at boot inside langchain.NewClient and exposed via Backend().
type LangChainProvider struct {
	client *langchain.Client
}

// NewLangChainProvider builds the provider from a constructed client.
func NewLangChainProvider(client *langchain.Client) *LangChainProvider {
	return &LangChainProvider{client: client}
}

func (p *LangChainProvider) Name() BotProviderName { return ProviderLangChain }

// Backend exposes the underlying LLM vendor name. Useful for log/audit
// surfaces that need to record which vendor served a request.
func (p *LangChainProvider) Backend() langchain.Backend { return p.client.Backend() }

// Chat invokes langchain.Client.Generate.
//
// The returned error is the typed MathError produced by mapLangChainError;
// callers (adapter.Chat) propagate it unchanged.
func (p *LangChainProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	out, err := p.client.Generate(ctx, toLangChainChat(req))
	if err != nil {
		return nil, mapLangChainError(ctx, p.client.Backend(), err)
	}
	return &ChatResult{
		Provider:     ProviderLangChain,
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

// Stream invokes langchain.Client.GenerateStream.
func (p *LangChainProvider) Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error) {
	out, err := p.client.GenerateStream(ctx, toLangChainChat(req), func(c langchain.StreamChunk) error {
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
		return nil, mapLangChainError(ctx, p.client.Backend(), err)
	}
	return &ChatResult{
		Provider:     ProviderLangChain,
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

// Embed invokes langchain.Client.Embed.
func (p *LangChainProvider) Embed(ctx context.Context, req EmbedRequest) (*EmbedResult, error) {
	out, err := p.client.Embed(ctx, langchain.EmbedRequest{
		Model:  req.Model,
		Inputs: req.Inputs,
	})
	if err != nil {
		return nil, mapLangChainError(ctx, p.client.Backend(), err)
	}
	return &EmbedResult{
		Provider: ProviderLangChain,
		Model:    out.Model,
		Vectors:  out.Vectors,
		Usage: Usage{
			PromptTokens:     out.Usage.PromptTokens,
			CompletionTokens: out.Usage.CompletionTokens,
			TotalTokens:      out.Usage.TotalTokens,
		},
	}, nil
}

func toLangChainChat(req ChatRequest) langchain.ChatRequest {
	msgs := make([]langchain.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, langchain.Message{
			Role:    toLangChainRole(m.Role),
			Content: m.Content,
			Name:    m.Name,
		})
	}
	return langchain.ChatRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stop:        req.Stop,
		JSONMode:    req.JSONMode,
	}
}

func toLangChainRole(r Role) langchain.Role {
	switch r {
	case RoleSystem:
		return langchain.RoleSystem
	case RoleAssistant:
		return langchain.RoleAssistant
	case RoleTool:
		return langchain.RoleTool
	default:
		return langchain.RoleUser
	}
}
