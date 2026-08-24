package eino

import (
	"errors"
	"testing"
	"time"
)

func validConfig() Config {
	return Config{
		Backend: BackendGoogleAI,
		APIKey:  "test-key",
		Model:   "gemini-1.5-flash",
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"valid googleai", func(c *Config) {}, false},
		{"valid openai", func(c *Config) { c.Backend = BackendOpenAI }, false},
		{"valid anthropic", func(c *Config) { c.Backend = BackendAnthropic }, false},
		{"valid ollama", func(c *Config) {
			c.Backend = BackendOllama
			c.APIKey = ""
			c.BaseURL = "http://localhost:11434"
		}, false},

		{"missing backend", func(c *Config) { c.Backend = "" }, true},
		{"unsupported backend", func(c *Config) { c.Backend = "grok" }, true},
		{"googleai missing api key", func(c *Config) { c.APIKey = "  " }, true},
		{"openai missing api key", func(c *Config) {
			c.Backend = BackendOpenAI
			c.APIKey = ""
		}, true},
		{"anthropic missing api key", func(c *Config) {
			c.Backend = BackendAnthropic
			c.APIKey = ""
		}, true},
		{"ollama missing base url", func(c *Config) {
			c.Backend = BackendOllama
			c.BaseURL = ""
		}, true},
		{"missing model", func(c *Config) { c.Model = "" }, true},
		{"bad base url scheme", func(c *Config) { c.BaseURL = "ftp://x" }, true},
		{"negative timeout", func(c *Config) { c.Timeout = -time.Second }, true},
		{"negative retries", func(c *Config) { c.MaxRetries = -1 }, true},
		{"negative retry delay", func(c *Config) { c.RetryDelay = -time.Millisecond }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.mutate(&cfg)

			err := cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Validate() = nil, want error")
				}
				if !errors.Is(err, ErrInvalidConfig) {
					t.Fatalf("Validate() error %v does not wrap ErrInvalidConfig", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestConfigApplyDefaults(t *testing.T) {
	cfg := validConfig()
	cfg.applyDefaults()

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
	}
	if cfg.MaxRetries != DefaultMaxRetries {
		t.Errorf("MaxRetries = %d, want %d", cfg.MaxRetries, DefaultMaxRetries)
	}
	if cfg.RetryDelay != DefaultRetryDelay {
		t.Errorf("RetryDelay = %v, want %v", cfg.RetryDelay, DefaultRetryDelay)
	}

	// Explicit values are preserved.
	cfg2 := validConfig()
	cfg2.Timeout = 5 * time.Second
	cfg2.MaxRetries = 7
	cfg2.RetryDelay = time.Second
	cfg2.applyDefaults()
	if cfg2.Timeout != 5*time.Second || cfg2.MaxRetries != 7 || cfg2.RetryDelay != time.Second {
		t.Errorf("applyDefaults overwrote explicit values: %+v", cfg2)
	}
}

// TestOpenAIExtraFields pins the only route to chat-completions `store`:
// eino-ext exposes no typed field for it, so it must ride in ExtraFields.
func TestOpenAIExtraFields(t *testing.T) {
	t.Run("off by default", func(t *testing.T) {
		if got := openAIExtraFields(validConfig()); got != nil {
			// nil, not an empty map: the request body must stay
			// byte-identical for deploys that never opt in.
			t.Errorf("openAIExtraFields() = %v, want nil", got)
		}
	})

	t.Run("store enabled", func(t *testing.T) {
		cfg := validConfig()
		cfg.Backend = BackendOpenAI
		cfg.Store = true

		got := openAIExtraFields(cfg)
		if len(got) != 1 || got["store"] != true {
			t.Fatalf("openAIExtraFields() = %v, want {store: true}", got)
		}
	})
}
