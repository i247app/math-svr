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
		adapter, err := NewFromConfig(context.Background(), config.BotConfig{BotProvider: provider})
		if err != nil {
			t.Errorf("provider=%q: err = %v, want nil", provider, err)
		}
		if adapter != nil {
			t.Errorf("provider=%q: adapter = %v, want nil", provider, adapter)
		}
	}
}

func TestNewFromConfigUnknownProvider(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{BotProvider: "skynet"})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigDefaultNotConfigured(t *testing.T) {
	// BOT_PROVIDER=eino but no BOT_EINO_BACKEND (and no langchain either):
	// nothing registers, SetDefault must fail loudly.
	_, err := NewFromConfig(context.Background(), config.BotConfig{BotProvider: "eino"})
	assertConfigInvalid(t, err)

	// BOT_PROVIDER=langchain while only eino is configured — the default
	// must point at a configured framework.
	_, err = NewFromConfig(context.Background(), config.BotConfig{
		BotProvider: "langchain",
		EinoBackend: "ollama",
		EinoBaseURL: "http://127.0.0.1:1",
		EinoModel:   "test-model",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigEinoInvalidBackend(t *testing.T) {
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		BotProvider: "eino",
		EinoBackend: "grok",
		EinoModel:   "m",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigEinoMissingCredential(t *testing.T) {
	// googleai requires an API key; Validate fails before any network I/O.
	_, err := NewFromConfig(context.Background(), config.BotConfig{
		BotProvider: "eino",
		EinoBackend: "googleai",
		EinoModel:   "gemini-1.5-flash",
	})
	assertConfigInvalid(t, err)
}

func TestNewFromConfigEinoOnlyRegistersAndDefaults(t *testing.T) {
	adapter, err := NewFromConfig(context.Background(), config.BotConfig{
		BotProvider: "eino",
		EinoBackend: "ollama",
		EinoBaseURL: "http://127.0.0.1:1",
		EinoModel:   "test-model",
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
