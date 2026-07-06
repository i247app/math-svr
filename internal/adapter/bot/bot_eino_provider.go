package bot

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/libs/eino"
)

// EinoProvider is the BotProvider implementation backed by libs/eino
// (cloudwego/eino + eino-ext). It owns nothing besides the wired
// *eino.Client and mirrors LangChainProvider's translation duties.
//
// The single EinoProvider can serve any of the eino-ext-backed LLM
// vendors (googleai, openai, anthropic, ollama) — the choice is fixed at
// boot inside eino.NewClient and exposed via Backend().
//
// Capability matrix: Chat and Stream are fully supported. Embed is NOT —
// it returns MathError(BOT_UNSUPPORTED_OP) per the BotProvider contract
// for optional capabilities (no eino-ext embedding component is wired).
type EinoProvider struct {
	client *eino.Client
}

// NewEinoProvider builds the provider from a constructed client.
func NewEinoProvider(client *eino.Client) *EinoProvider {
	return &EinoProvider{client: client}
}

func (p *EinoProvider) Name() BotProviderName { return ProviderEino }

// Backend exposes the underlying LLM vendor name. Useful for log/audit
// surfaces that need to record which vendor served a request.
func (p *EinoProvider) Backend() eino.Backend { return p.client.Backend() }

// Chat invokes eino.Client.Generate.
//
// The returned error is the typed MathError produced by mapEinoError;
// callers (adapter.Chat) propagate it unchanged.
func (p *EinoProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	out, err := p.client.Generate(ctx, toEinoChat(req))
	if err != nil {
		return nil, mapEinoError(ctx, p.client.Backend(), err)
	}
	return &ChatResult{
		Provider:     ProviderEino,
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

// Stream invokes eino.Client.GenerateStream.
func (p *EinoProvider) Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error) {
	out, err := p.client.GenerateStream(ctx, toEinoChat(req), func(c eino.StreamChunk) error {
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
		return nil, mapEinoError(ctx, p.client.Backend(), err)
	}
	return &ChatResult{
		Provider:     ProviderEino,
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

// Embed reports the capability as unsupported. The eino provider does not
// wire an embedding component today; deploys that need embeddings should
// route Embed via the langchain provider (EmbedVia) or extend libs/eino.
func (p *EinoProvider) Embed(ctx context.Context, _ EmbedRequest) (*EmbedResult, error) {
	return nil, errs.NewError(ctx, status.BOT_UNSUPPORTED_OP,
		map[string]any{
			"provider": string(ProviderEino),
			"backend":  string(p.client.Backend()),
		},
		errors.New("eino provider does not support embeddings"))
}

func toEinoChat(req ChatRequest) eino.ChatRequest {
	msgs := make([]eino.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, eino.Message{
			Role:    toEinoRole(m.Role),
			Content: m.Content,
			Name:    m.Name,
		})
	}
	return eino.ChatRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stop:        req.Stop,
		JSONMode:    req.JSONMode,
	}
}

func toEinoRole(r Role) eino.Role {
	switch r {
	case RoleSystem:
		return eino.RoleSystem
	case RoleAssistant:
		return eino.RoleAssistant
	case RoleTool:
		return eino.RoleTool
	default:
		return eino.RoleUser
	}
}
