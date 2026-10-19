package llm

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/libs/deepseek"
	"math-ai.com/math-ai/internal/libs/gemini"
	"math-ai.com/math-ai/internal/libs/openai"
	"math-ai.com/math-ai/internal/libs/openrouter"
)

// NewFromConfig builds a fully-wired *Adapter from cfg, or returns
// (nil, nil) when the LLM adapter is intentionally disabled in this
// deploy.
//
// Behaviour matrix:
//
//	cfg.LLMProvider == "" or "disabled" → (nil, nil); boot continues.
//	cfg.LLMProvider == "openrouter"|"openai"|"gemini"|"deepseek" →
//	    every configured provider is registered:
//	      - openrouter when cfg.OpenRouterAPIKey != ""
//	      - openai     when cfg.OpenAIAPIKey     != ""
//	      - gemini     when cfg.GeminiAPIKey     != ""
//	      - deepseek   when cfg.DeepSeekAPIKey   != ""
//	    then cfg.LLMProvider is set as the default. Callers can still
//	    route per-call to a specific provider via ChatVia / StreamVia.
//	cfg.LLMProvider == anything else    → MathError(LLM_CONFIG_INVALID).
//
// Every provider keys off its API key: OpenRouter's model id picks the
// vendor, and each direct client talks to one vendor.
//
// Naming a default whose provider is not configured (e.g.
// LLM_PROVIDER=gemini without LLM_GEMINI_API_KEY) is a deploy mistake and
// fails with LLM_CONFIG_INVALID.
//
// Disabled returns nil rather than an error so dev profiles without
// LLM credentials can boot. Module services that consume the adapter
// must nil-guard, the same way they do with res.SMSProvider in local dev.
//
// Errors from libs/openrouter, libs/openai, libs/gemini and libs/deepseek
// are translated here:
//
//	ErrInvalidConfig → MathError(LLM_CONFIG_INVALID, {"reason": ...})
//	anything else    → MathError(LLM_CONNECT_FAILED) wrapping the cause
func NewFromConfig(ctx context.Context, cfg config.LLMConfig) (*Adapter, error) {
	log := logger.From(ctx)

	switch cfg.DefaultLLMProvider {
	case "", "disabled":
		return nil, nil
	case string(ProviderOpenRouter), string(ProviderOpenAI),
		string(ProviderGemini), string(ProviderDeepSeek):
		// recognised — continue below
	default:
		return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID,
			map[string]any{"provider": cfg.DefaultLLMProvider}, nil)
	}

	adapter := NewAdapter()

	if cfg.OpenRouterAPIKey != "" {
		client, err := openrouter.NewClient(ctx, openrouter.Config{
			APIKey:        cfg.OpenRouterAPIKey,
			BaseURL:       cfg.OpenRouterBaseURL,
			Model:         cfg.OpenRouterModel,
			SiteURL:       cfg.OpenRouterSiteURL,
			AppTitle:      cfg.OpenRouterAppTitle,
			Temperature:   cfg.OpenRouterTemperature,
			TopP:          cfg.OpenRouterTopP,
			MaxTokens:     cfg.OpenRouterMaxTokens,
			Timeout:       cfg.Timeout,
			MaxRetries:    cfg.MaxRetries,
			RetryDelay:    cfg.RetryDelay,
			RequireAtBoot: cfg.RequireAtBoot,
		})
		if err != nil {
			if errors.Is(err, openrouter.ErrInvalidConfig) {
				return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			log.Warnf("openrouter: client not ready: %v", err)
			return nil, errs.NewError(ctx, status.LLM_CONNECT_FAILED,
				map[string]any{"model": cfg.OpenRouterModel}, err)
		}

		adapter.Register(NewOpenRouterProvider(client))
	}

	if cfg.OpenAIAPIKey != "" {
		client, err := openai.NewClient(ctx, openai.Config{
			APIKey:          cfg.OpenAIAPIKey,
			BaseURL:         cfg.OpenAIBaseURL,
			Model:           cfg.OpenAIModel,
			EmbedModel:      cfg.OpenAIEmbedModel,
			Organization:    cfg.OpenAIOrganization,
			Project:         cfg.OpenAIProject,
			Store:           cfg.OpenAIStore,
			Temperature:     cfg.OpenAITemperature,
			TopP:            cfg.OpenAITopP,
			MaxTokens:       cfg.OpenAIMaxTokens,
			ReasoningEffort: cfg.OpenAIReasoningEffort,
			Timeout:         cfg.Timeout,
			MaxRetries:      cfg.MaxRetries,
			RetryDelay:      cfg.RetryDelay,
			RequireAtBoot:   cfg.RequireAtBoot,
		})
		if err != nil {
			if errors.Is(err, openai.ErrInvalidConfig) {
				return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			log.Warnf("openai: client not ready: %v", err)
			return nil, errs.NewError(ctx, status.LLM_CONNECT_FAILED,
				map[string]any{"model": cfg.OpenAIModel}, err)
		}

		adapter.Register(NewOpenAIProvider(client))
	}

	if cfg.GeminiAPIKey != "" {
		client, err := gemini.NewClient(ctx, gemini.Config{
			APIKey:        cfg.GeminiAPIKey,
			BaseURL:       cfg.GeminiBaseURL,
			Model:         cfg.GeminiModel,
			EmbedModel:    cfg.GeminiEmbedModel,
			Temperature:   cfg.GeminiTemperature,
			TopP:          cfg.GeminiTopP,
			MaxTokens:     cfg.GeminiMaxTokens,
			Timeout:       cfg.Timeout,
			MaxRetries:    cfg.MaxRetries,
			RetryDelay:    cfg.RetryDelay,
			RequireAtBoot: cfg.RequireAtBoot,
		})
		if err != nil {
			if errors.Is(err, gemini.ErrInvalidConfig) {
				return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			log.Warnf("gemini: client not ready: %v", err)
			return nil, errs.NewError(ctx, status.LLM_CONNECT_FAILED,
				map[string]any{"model": cfg.GeminiModel}, err)
		}

		adapter.Register(NewGeminiProvider(client))
	}

	if cfg.DeepSeekAPIKey != "" {
		client, err := deepseek.NewClient(ctx, deepseek.Config{
			APIKey:          cfg.DeepSeekAPIKey,
			BaseURL:         cfg.DeepSeekBaseURL,
			Model:           cfg.DeepSeekModel,
			Thinking:        cfg.DeepSeekThinking,
			ReasoningEffort: cfg.DeepSeekReasoningEffort,
			Temperature:     cfg.DeepSeekTemperature,
			TopP:            cfg.DeepSeekTopP,
			MaxTokens:       cfg.DeepSeekMaxTokens,
			Timeout:         cfg.Timeout,
			MaxRetries:      cfg.MaxRetries,
			RetryDelay:      cfg.RetryDelay,
			RequireAtBoot:   cfg.RequireAtBoot,
		})
		if err != nil {
			if errors.Is(err, deepseek.ErrInvalidConfig) {
				return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			log.Warnf("deepseek: client not ready: %v", err)
			return nil, errs.NewError(ctx, status.LLM_CONNECT_FAILED,
				map[string]any{"model": cfg.DeepSeekModel}, err)
		}

		adapter.Register(NewDeepSeekProvider(client))
	}

	if err := adapter.SetDefault(LLMProviderName(cfg.DefaultLLMProvider)); err != nil {
		return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID,
			map[string]any{
				"provider": cfg.DefaultLLMProvider,
				"reason":   "default provider is not configured (missing LLM_<PROVIDER>_API_KEY?)",
			}, err)
	}

	return adapter, nil
}
