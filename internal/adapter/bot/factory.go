package bot

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/libs/eino"
	"math-ai.com/math-ai/internal/libs/gemini"
	"math-ai.com/math-ai/internal/libs/langchain"
	"math-ai.com/math-ai/internal/libs/openai"
	"math-ai.com/math-ai/internal/libs/openrouter"
)

// NewFromConfig builds a fully-wired *Adapter from cfg, or returns
// (nil, nil) when the bot adapter is intentionally disabled in this
// deploy.
//
// Behaviour matrix:
//
//	cfg.BotProvider == "" or "disabled" → (nil, nil); boot continues.
//	cfg.BotProvider == "langchain"|"eino"|"openrouter"|"openai"|"gemini" →
//	    every configured framework is registered:
//	      - langchain  when cfg.LangChainBackend != ""
//	      - eino       when cfg.EinoBackend      != ""
//	      - openrouter when cfg.OpenRouterAPIKey != ""
//	      - openai     when cfg.OpenAIAPIKey     != ""
//	      - gemini     when cfg.GeminiAPIKey     != ""
//	    then cfg.BotProvider is set as the default. Callers can still
//	    route per-call to a specific provider via ChatVia / StreamVia.
//	cfg.BotProvider == anything else    → MathError(BOT_CONFIG_INVALID).
//
// openrouter, openai and gemini key off their API key rather than a
// backend name because none has a backend selector: OpenRouter's model id
// picks the vendor, and the two direct clients each talk to one vendor.
//
// Naming a default whose framework is not configured (e.g.
// BOT_PROVIDER=eino without BOT_EINO_BACKEND) is a deploy mistake and
// fails with BOT_CONFIG_INVALID.
//
// Disabled returns nil rather than an error so dev profiles without
// LLM credentials can boot. Module services that consume the adapter
// must nil-guard, the same way they do with res.SMSProvider in local dev.
//
// Errors from libs/langchain, libs/eino, libs/openrouter, libs/openai and
// libs/gemini are translated here:
//
//	ErrInvalidConfig → MathError(BOT_CONFIG_INVALID, {"reason": ...})
//	anything else    → MathError(BOT_CONNECT_FAILED) wrapping the cause
func NewFromConfig(ctx context.Context, cfg config.BotConfig) (*Adapter, error) {
	log := logger.From(ctx)

	switch cfg.DefaultBotProvider {
	case "", "disabled":
		return nil, nil
	case string(ProviderLangChain), string(ProviderEino),
		string(ProviderOpenRouter), string(ProviderOpenAI), string(ProviderGemini):
		// recognised — continue below
	default:
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID,
			map[string]any{"provider": cfg.DefaultBotProvider}, nil)
	}

	adapter := NewAdapter()

	if cfg.LangChainBackend != "" {
		client, err := langchain.NewClient(ctx, langchain.Config{
			Backend:       langchain.Backend(cfg.LangChainBackend),
			APIKey:        cfg.LangChainAPIKey,
			BaseURL:       cfg.LangChainBaseURL,
			Model:         cfg.LangChainModel,
			EmbedModel:    cfg.LangChainEmbedModel,
			Temperature:   cfg.LangChainTemperature,
			TopP:          cfg.LangChainTopP,
			MaxTokens:     cfg.LangChainMaxTokens,
			Timeout:       cfg.Timeout,
			MaxRetries:    cfg.MaxRetries,
			RetryDelay:    cfg.RetryDelay,
			RequireAtBoot: cfg.RequireAtBoot,
		})
		if err != nil {
			if errors.Is(err, langchain.ErrInvalidConfig) {
				return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			return nil, errs.NewError(ctx, status.BOT_CONNECT_FAILED,
				map[string]any{"backend": cfg.LangChainBackend}, err)
		}
		adapter.Register(NewLangChainProvider(client))
	}

	if cfg.EinoBackend != "" {
		client, err := eino.NewClient(ctx, eino.Config{
			Backend:       eino.Backend(cfg.EinoBackend),
			APIKey:        cfg.EinoAPIKey,
			BaseURL:       cfg.EinoBaseURL,
			Model:         cfg.EinoModel,
			Temperature:   cfg.EinoTemperature,
			TopP:          cfg.EinoTopP,
			MaxTokens:     cfg.EinoMaxTokens,
			Store:         cfg.EinoStore,
			Timeout:       cfg.Timeout,
			MaxRetries:    cfg.MaxRetries,
			RetryDelay:    cfg.RetryDelay,
			RequireAtBoot: cfg.RequireAtBoot,
		})

		if err != nil {
			log.Error("eino: client not rceady %v", err)
			if errors.Is(err, eino.ErrInvalidConfig) {
				return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			return nil, errs.NewError(ctx, status.BOT_CONNECT_FAILED,
				map[string]any{"backend": cfg.EinoBackend}, err)
		}

		adapter.Register(NewEinoProvider(client))
	}

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
				return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			log.Warnf("openrouter: client not ready: %v", err)
			return nil, errs.NewError(ctx, status.BOT_CONNECT_FAILED,
				map[string]any{"model": cfg.OpenRouterModel}, err)
		}

		adapter.Register(NewOpenRouterProvider(client))
	}

	if cfg.OpenAIAPIKey != "" {
		client, err := openai.NewClient(ctx, openai.Config{
			APIKey:        cfg.OpenAIAPIKey,
			BaseURL:       cfg.OpenAIBaseURL,
			Model:         cfg.OpenAIModel,
			EmbedModel:    cfg.OpenAIEmbedModel,
			Organization:  cfg.OpenAIOrganization,
			Project:       cfg.OpenAIProject,
			Store:         cfg.OpenAIStore,
			Temperature:   cfg.OpenAITemperature,
			TopP:          cfg.OpenAITopP,
			MaxTokens:     cfg.OpenAIMaxTokens,
			Timeout:       cfg.Timeout,
			MaxRetries:    cfg.MaxRetries,
			RetryDelay:    cfg.RetryDelay,
			RequireAtBoot: cfg.RequireAtBoot,
		})
		if err != nil {
			if errors.Is(err, openai.ErrInvalidConfig) {
				return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			log.Warnf("openai: client not ready: %v", err)
			return nil, errs.NewError(ctx, status.BOT_CONNECT_FAILED,
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
				return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			log.Warnf("gemini: client not ready: %v", err)
			return nil, errs.NewError(ctx, status.BOT_CONNECT_FAILED,
				map[string]any{"model": cfg.GeminiModel}, err)
		}

		adapter.Register(NewGeminiProvider(client))
	}

	if err := adapter.SetDefault(BotProviderName(cfg.DefaultBotProvider)); err != nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID,
			map[string]any{
				"provider": cfg.DefaultBotProvider,
				"reason":   "default provider is not configured (missing BOT_<PROVIDER>_BACKEND?)",
			}, err)
	}

	return adapter, nil
}
