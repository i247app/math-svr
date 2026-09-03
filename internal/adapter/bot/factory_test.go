package bot

import (
	"context"
	"testing"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
)

// Factory tests stay hermetic: every case either short-circuits before any
// vendor construction or uses the ollama backend, whose eino-ext component
// builds a pure in-process client (no network I/O at construction; the
// boot probe is skipped because RequireAtBoot defaults to false).

func assertConfigInvalid(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("NewFromConfig() error = nil, want BOT_CONFIG_INVALID")
	}
	mErr, ok := errs.IsMathError(err)
	if !ok {
		t.Fatalf("error %v is not a MathError", err)
	}
	if mErr.GetStatusCode() != status.BOT_CONFIG_INVALID {
		t.Fatalf("status = %d, want BOT_CONFIG_INVALID", mErr.GetStatusCode())
	}
}

func TestNewFromConfigDisabled(t *testing.T) {
	for _, provider := range []string{"", "disabled"} {
		adapter, err := NewFromConfig(context.Background(), config.BotConfig{DefaultBotProvider: provider})
		if err != nil {
			t.Errorf("provider=%q: err = %v, want nil", provider, err)
		}
		if adapter != nil {
			t.Errorf("provider=%q: adapter = %v, want nil", provider, adapter)
		}
	}
}

func TestNewFromConfigUnknownProvider(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{DefaultBotProvider: "skynet"})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigDefaultNotConfigured(t *testing.T) {
	// BOT_PROVIDER=eino but no BOT_EINO_BACKEND (and no langchain either):
	// nothing registers, SetDefault must fail loudly.
	_, err := NewFromConfig(context.Background(), config.BotConfig{DefaultBotProvider: "eino"})
	assertConfigInvalid(t, err)

	// BOT_PROVIDER=langchain while only eino is configured — the default
	// must point at a configured framework.
	_, err = NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "langchain",
		EinoBackend:        "ollama",
		EinoBaseURL:        "http://127.0.0.1:1",
		EinoModel:          "test-model",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigEinoInvalidBackend(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "eino",
		EinoBackend:        "grok",
		EinoModel:          "m",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigEinoMissingCredential(t *testing.T) {
	// googleai requires an API key; Validate fails before any network I/O.
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "eino",
		EinoBackend:        "googleai",
		EinoModel:          "gemini-1.5-flash",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigEinoOnlyRegistersAndDefaults(t *testing.T) {
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "eino",
		EinoBackend:        "ollama",
		EinoBaseURL:        "http://127.0.0.1:1",
		EinoModel:          "test-model",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if adapter == nil {
		t.Fatal("adapter = nil, want registered adapter")
	}
	if adapter.DefaultName() != ProviderEino {
		t.Errorf("DefaultName() = %s, want eino", adapter.DefaultName())
	}
	if !adapter.Has(ProviderEino) {
		t.Error("eino provider not registered")
	}
	if adapter.Has(ProviderLangChain) {
		t.Error("langchain must NOT register when its backend key is empty")
	}
}

func TestNewFromConfigOpenRouterRegistersAndDefaults(t *testing.T) {
	// The openrouter provider builds no vendor SDK and RequireAtBoot
	// defaults to false, so construction performs no network I/O.
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "openrouter",
		OpenRouterAPIKey:   "test-key",
		OpenRouterModel:    "openai/gpt-4o-mini",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if adapter == nil {
		t.Fatal("adapter = nil, want registered adapter")
	}
	if adapter.DefaultName() != ProviderOpenRouter {
		t.Errorf("DefaultName() = %s, want openrouter", adapter.DefaultName())
	}
	if !adapter.Has(ProviderOpenRouter) {
		t.Error("openrouter provider not registered")
	}
	if adapter.Has(ProviderEino) || adapter.Has(ProviderLangChain) {
		t.Error("only the configured framework may register")
	}
}

func TestNewFromConfigOpenRouterMissingModel(t *testing.T) {
	// There is deliberately no default model: an operator who supplies a
	// key but no model must be told, not silently billed for a guess.
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "openrouter",
		OpenRouterAPIKey:   "test-key",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigOpenRouterMissingCredential(t *testing.T) {
	// BOT_PROVIDER=openrouter without BOT_OPENROUTER_API_KEY registers
	// nothing, so SetDefault must fail loudly rather than boot a bot-less
	// adapter that 500s on the first quiz.
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "openrouter",
		OpenRouterModel:    "openai/gpt-4o-mini",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigOpenRouterAlongsideEino(t *testing.T) {
	// Both frameworks configured, eino named as default: openrouter is
	// still registered and reachable per-call via ChatVia.
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "eino",
		EinoBackend:        "ollama",
		EinoBaseURL:        "http://127.0.0.1:1",
		EinoModel:          "test-model",
		OpenRouterAPIKey:   "test-key",
		OpenRouterModel:    "openai/gpt-4o-mini",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if adapter.DefaultName() != ProviderEino {
		t.Errorf("DefaultName() = %s, want eino", adapter.DefaultName())
	}
	if !adapter.Has(ProviderOpenRouter) {
		t.Error("openrouter must register whenever its API key is set")
	}
}

func TestNewFromConfigOpenAIRegistersAndDefaults(t *testing.T) {
	// The direct openai provider builds no vendor SDK and RequireAtBoot
	// defaults to false, so construction performs no network I/O.
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "openai",
		OpenAIAPIKey:       "test-key",
		OpenAIModel:        "gpt-4.1",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if adapter == nil {
		t.Fatal("adapter = nil, want registered adapter")
	}
	if adapter.DefaultName() != ProviderOpenAI {
		t.Errorf("DefaultName() = %s, want openai", adapter.DefaultName())
	}
	if !adapter.Has(ProviderOpenAI) {
		t.Error("openai provider not registered")
	}
	if adapter.Has(ProviderEino) || adapter.Has(ProviderLangChain) || adapter.Has(ProviderOpenRouter) {
		t.Error("only the configured framework may register")
	}
}

func TestNewFromConfigOpenAIMissingModel(t *testing.T) {
	// No default model on purpose: an operator who supplies a key but no
	// model must be told, not silently billed for a guess.
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "openai",
		OpenAIAPIKey:       "test-key",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigOpenAIMissingCredential(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "openai",
		OpenAIModel:        "gpt-4.1",
	})
	assertConfigInvalid(t, err)
}

// TestNewFromConfigAllFourProviders proves the four coexist on one adapter
// and that BOT_PROVIDER only picks the DEFAULT — every configured provider
// stays reachable per-call through ChatVia / StreamVia / EmbedVia.
func TestNewFromConfigAllFourProviders(t *testing.T) {
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "openai",

		EinoBackend: "ollama",
		EinoBaseURL: "http://127.0.0.1:1",
		EinoModel:   "test-model",

		OpenRouterAPIKey: "test-key",
		OpenRouterModel:  "openai/gpt-4o-mini",

		OpenAIAPIKey: "test-key",
		OpenAIModel:  "gpt-4.1",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if adapter.DefaultName() != ProviderOpenAI {
		t.Errorf("DefaultName() = %s, want openai", adapter.DefaultName())
	}
	for _, name := range []BotProviderName{ProviderEino, ProviderOpenRouter, ProviderOpenAI} {
		if !adapter.Has(name) {
			t.Errorf("provider %s not registered", name)
		}
	}
}

func TestNewFromConfigGeminiRegistersAndDefaults(t *testing.T) {
	// The direct gemini provider builds no vendor SDK and RequireAtBoot
	// defaults to false, so construction performs no network I/O.
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "gemini",
		GeminiAPIKey:       "test-key",
		GeminiModel:        "gemini-2.0-flash",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if adapter == nil {
		t.Fatal("adapter = nil, want registered adapter")
	}
	if adapter.DefaultName() != ProviderGemini {
		t.Errorf("DefaultName() = %s, want gemini", adapter.DefaultName())
	}
	if !adapter.Has(ProviderGemini) {
		t.Error("gemini provider not registered")
	}
	if adapter.Has(ProviderEino) || adapter.Has(ProviderLangChain) ||
		adapter.Has(ProviderOpenRouter) || adapter.Has(ProviderOpenAI) {
		t.Error("only the configured framework may register")
	}
}

func TestNewFromConfigGeminiMissingModel(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "gemini",
		GeminiAPIKey:       "test-key",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigGeminiMissingCredential(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "gemini",
		GeminiModel:        "gemini-2.0-flash",
	})
	assertConfigInvalid(t, err)
}

// TestNewFromConfigAllFiveProviders proves the five coexist on one adapter
// and that BOT_PROVIDER only picks the DEFAULT — every configured provider
// stays reachable per-call through ChatVia / StreamVia / EmbedVia.
func TestNewFromConfigAllFiveProviders(t *testing.T) {
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "gemini",

		EinoBackend: "ollama",
		EinoBaseURL: "http://127.0.0.1:1",
		EinoModel:   "test-model",

		OpenRouterAPIKey: "test-key",
		OpenRouterModel:  "openai/gpt-4o-mini",

		OpenAIAPIKey: "test-key",
		OpenAIModel:  "gpt-4.1",

		GeminiAPIKey: "test-key",
		GeminiModel:  "gemini-2.0-flash",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if adapter.DefaultName() != ProviderGemini {
		t.Errorf("DefaultName() = %s, want gemini", adapter.DefaultName())
	}
	for _, name := range []BotProviderName{
		ProviderEino, ProviderOpenRouter, ProviderOpenAI, ProviderGemini,
	} {
		if !adapter.Has(name) {
			t.Errorf("provider %s not registered", name)
		}
	}
}

func TestNewFromConfigDeepSeekRegistersAndDefaults(t *testing.T) {
	// The direct deepseek provider builds no vendor SDK and RequireAtBoot
	// defaults to false, so construction performs no network I/O.
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "deepseek",
		DeepSeekAPIKey:     "test-key",
		DeepSeekModel:      "deepseek-v4-flash",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if adapter == nil {
		t.Fatal("adapter = nil, want registered adapter")
	}
	if adapter.DefaultName() != ProviderDeepSeek {
		t.Errorf("DefaultName() = %s, want deepseek", adapter.DefaultName())
	}
	if !adapter.Has(ProviderDeepSeek) {
		t.Error("deepseek provider not registered")
	}
}

func TestNewFromConfigDeepSeekMissingModel(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "deepseek",
		DeepSeekAPIKey:     "test-key",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigDeepSeekMissingCredential(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "deepseek",
		DeepSeekModel:      "deepseek-v4-flash",
	})
	assertConfigInvalid(t, err)
}

// TestNewFromConfigDeepSeekBadThinkingValue: an unknown enum would be a
// 400 on every single call, so it must fail at boot instead.
func TestNewFromConfigDeepSeekBadThinkingValue(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "deepseek",
		DeepSeekAPIKey:     "test-key",
		DeepSeekModel:      "deepseek-v4-flash",
		DeepSeekThinking:   "on",
	})
	assertConfigInvalid(t, err)
}

// TestNewFromConfigAllSixProviders proves the six coexist on one adapter
// and that BOT_PROVIDER only picks the DEFAULT — every configured provider
// stays reachable per-call through ChatVia / StreamVia / EmbedVia.
func TestNewFromConfigAllSixProviders(t *testing.T) {
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		DefaultBotProvider: "deepseek",

		EinoBackend: "ollama",
		EinoBaseURL: "http://127.0.0.1:1",
		EinoModel:   "test-model",

		OpenRouterAPIKey: "test-key",
		OpenRouterModel:  "openai/gpt-4o-mini",

		OpenAIAPIKey: "test-key",
		OpenAIModel:  "gpt-4.1",

		GeminiAPIKey: "test-key",
		GeminiModel:  "gemini-2.0-flash",

		DeepSeekAPIKey: "test-key",
		DeepSeekModel:  "deepseek-v4-flash",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if adapter.DefaultName() != ProviderDeepSeek {
		t.Errorf("DefaultName() = %s, want deepseek", adapter.DefaultName())
	}
	for _, name := range []BotProviderName{
		ProviderEino, ProviderOpenRouter, ProviderOpenAI, ProviderGemini, ProviderDeepSeek,
	} {
		if !adapter.Has(name) {
			t.Errorf("provider %s not registered", name)
		}
	}
}
