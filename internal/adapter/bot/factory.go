package bot

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/libs/langchain"
)

// NewFromConfig builds a fully-wired *Adapter from cfg, or returns
// (nil, nil) when the bot adapter is intentionally disabled in this
// deploy.
//
// Behaviour matrix:
//
//	cfg.BotProvider == "" or "disabled" → (nil, nil); boot continues.
//	cfg.BotProvider == "langchain"      → langchain provider, registered + defaulted.
//	cfg.BotProvider == anything else    → MathError(BOT_CONFIG_INVALID).
//
// Disabled returns nil rather than an error so dev profiles without
// LLM credentials can boot. Module services that consume the adapter
// must nil-guard, the same way they do with res.SMSProvider in local dev.
//
// Errors from libs/langchain are translated here:
//
//	ErrInvalidConfig → MathError(BOT_CONFIG_INVALID, {"reason": ...})
//	anything else    → MathError(BOT_CONNECT_FAILED) wrapping the cause
func NewFromConfig(ctx context.Context, cfg config.BotConfig) (*Adapter, error) {
	switch cfg.BotProvider {
	case "", "disabled":
		return nil, nil

	case string(ProviderLangChain):
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

		adapter := NewAdapter()
		adapter.Register(NewLangChainProvider(client))
		return adapter, nil

	default:
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID,
			map[string]any{"provider": cfg.BotProvider}, nil)
	}
}
