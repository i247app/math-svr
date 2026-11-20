package llm

import (
	"context"
	"slices"
	"strings"
)

// ListModels implementations, one per vendor: each calls its client's
// model-list endpoint and maps the error the way Chat does.

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	models, err := p.client.ListModels(ctx)
	if err != nil {
		return nil, mapOpenAIError(ctx, err)
	}
	out := make([]ModelInfo, 0, len(models))
	for _, m := range models {
		modelInfo := ModelInfo{
			ID:      m.ID,
			OwnedBy: m.OwnedBy,
		}
		out = append(out, modelInfo)
	}
	return out, nil
}

func (p *DeepSeekProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	models, err := p.client.ListModels(ctx)
	if err != nil {
		return nil, mapDeepSeekError(ctx, err)
	}
	out := make([]ModelInfo, 0, len(models))
	for _, m := range models {
		modelInfo := ModelInfo{
			ID:      m.ID,
			OwnedBy: m.OwnedBy,
		}

		out = append(out, modelInfo)
	}
	return out, nil
}

func (p *GeminiProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	models, err := p.client.ListModels(ctx)
	if err != nil {
		return nil, mapGeminiError(ctx, err)
	}
	out := make([]ModelInfo, 0, len(models))
	for _, m := range models {
		modelInfo := ModelInfo{
			ID:               m.ID,
			Name:             m.Name,
			ContextLength:    m.InputTokenLimit,
			SupportedMethods: m.SupportedMethods,
		}

		out = append(out, modelInfo)
	}
	return out, nil
}

func (p *OpenRouterProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	models, err := p.client.ListModels(ctx)
	if err != nil {
		return nil, mapOpenRouterError(ctx, err)
	}
	out := make([]ModelInfo, 0, len(models))
	for _, m := range models {
		modelInfo := ModelInfo{
			ID:               m.ID,
			Name:             m.Name,
			ContextLength:    m.ContextLength,
			InputModalities:  m.InputModalities,
			OutputModalities: m.OutputModalities,
		}

		out = append(out, modelInfo)
	}
	return out, nil
}

// Compile-time proof every vendor keeps the capability.
var (
	_ ModelLister = (*OpenAIProvider)(nil)
	_ ModelLister = (*DeepSeekProvider)(nil)
	_ ModelLister = (*GeminiProvider)(nil)
	_ ModelLister = (*OpenRouterProvider)(nil)
)

// openAINonChatMarkers are id fragments of OpenAI models that are not chat
// completions models. OpenAI's list has no capability flag, so this is a
// name heuristic and can lag a new model family.
var openAINonChatMarkers = []string{
	"embedding", "tts", "whisper", "dall-e", "image", "moderation", "transcribe", "realtime",
	"babbage", "davinci",
}

// IsChatModel reports whether a listed model can serve a chat call, by the
// hints its provider gives:
//   - gemini: "generateContent" in SupportedMethods;
//   - openrouter: "text" in OutputModalities (kept when the vendor gave none);
//   - openai: the id names no known non-chat family (openAINonChatMarkers);
//   - deepseek and any other provider: every model (chat only today).
func IsChatModel(provider LLMProviderName, m ModelInfo) bool {
	switch provider {
	case ProviderGemini:
		return slices.Contains(m.SupportedMethods, "generateContent")
	case ProviderOpenRouter:
		return len(m.OutputModalities) == 0 || slices.Contains(m.OutputModalities, "text")
	case ProviderOpenAI:
		id := strings.ToLower(m.ID)
		for _, marker := range openAINonChatMarkers {
			if strings.Contains(id, marker) {
				return false
			}
		}
		return true
	}
	return true
}
