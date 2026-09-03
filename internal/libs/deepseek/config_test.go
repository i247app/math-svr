package deepseek

import (
	"errors"
	"testing"
	"time"
)

func validConfig() Config {
	return Config{
		APIKey: "test-key",
		Model:  "deepseek-v4-flash",
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"valid", func(c *Config) {}, false},
		{"valid with thinking enabled", func(c *Config) {
			c.Thinking = ThinkingEnabled
			c.ReasoningEffort = ReasoningEffortHigh
		}, false},
		{"valid with thinking disabled", func(c *Config) { c.Thinking = ThinkingDisabled }, false},
		{"valid with beta base url", func(c *Config) { c.BaseURL = "https://api.deepseek.com/beta" }, false},

		{"missing api key", func(c *Config) { c.APIKey = "  " }, true},
		{"missing model", func(c *Config) { c.Model = "" }, true},
		{"bad base url scheme", func(c *Config) { c.BaseURL = "ftp://x" }, true},
		// A typo here is a 400 on every single call, so it must fail at boot.
		{"bad thinking value", func(c *Config) { c.Thinking = "on" }, true},
		{"bad reasoning effort", func(c *Config) { c.ReasoningEffort = "medium" }, true},
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

	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, DefaultBaseURL)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
	}
	if cfg.MaxRetries != DefaultMaxRetries {
		t.Errorf("MaxRetries = %d, want %d", cfg.MaxRetries, DefaultMaxRetries)
	}
	if cfg.RetryDelay != DefaultRetryDelay {
		t.Errorf("RetryDelay = %v, want %v", cfg.RetryDelay, DefaultRetryDelay)
	}
}

// TestClassifyInsufficientBalance: DeepSeek uses HTTP 402 for an empty
// balance, the way OpenRouter does — not OpenAI's 429-plus-code. It must
// never be retried and must reach the caller as a config problem.
func TestClassifyInsufficientBalance(t *testing.T) {
	err := classifyAPIError(&APIError{HTTPStatus: 402, Message: "Insufficient Balance"})

	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("error = %v, want ErrInsufficientBalance", err)
	}
	if !IsConfigError(err) {
		t.Error("402 must classify as a config error (needs an operator to top up)")
	}
	if IsRateLimited(err) {
		t.Error("402 must not classify as rate-limited")
	}
	if _, ok := retryDelayFor(&APIError{HTTPStatus: 402}, time.Second); ok {
		t.Error("402 must never be retried — the balance will not refill mid-call")
	}
}

func TestClassifyStatuses(t *testing.T) {
	tests := []struct {
		name   string
		api    *APIError
		wantIs error
	}{
		{"429 throttling", &APIError{HTTPStatus: 429}, ErrRateLimited},
		{"503 overloaded", &APIError{HTTPStatus: 503}, ErrServerOverloaded},
		{
			"400 context length",
			&APIError{HTTPStatus: 400, Message: "This model's maximum context length is 65536 tokens"},
			ErrContextTooLarge,
		},
		{
			// 422 is DeepSeek's "invalid parameters"; an oversized prompt
			// can land on either status.
			"422 context length",
			&APIError{HTTPStatus: 422, Message: "input exceeds the maximum length"},
			ErrContextTooLarge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := classifyAPIError(tt.api); !errors.Is(err, tt.wantIs) {
				t.Errorf("error = %v, want %v", err, tt.wantIs)
			}
		})
	}

	plain := classifyAPIError(&APIError{HTTPStatus: 400, Message: "messages must be an array"})
	if errors.Is(plain, ErrContextTooLarge) {
		t.Error("a plain 400 must not be classified as context-too-large")
	}
}

func TestRetryDelayFor(t *testing.T) {
	base := 500 * time.Millisecond
	tests := []struct {
		name    string
		api     *APIError
		wantOK  bool
		wantDur time.Duration
	}{
		{"nil", nil, false, 0},
		{"429 no retry-after", &APIError{HTTPStatus: 429}, false, 0},
		{"429 with retry-after", &APIError{HTTPStatus: 429, RetryAfter: 2 * time.Second}, true, 2 * time.Second},
		{"429 retry-after too long", &APIError{HTTPStatus: 429, RetryAfter: maxRetryAfter + time.Second}, false, 0},
		{"500", &APIError{HTTPStatus: 500}, true, base},
		{"503", &APIError{HTTPStatus: 503}, true, base},
		{"401", &APIError{HTTPStatus: 401}, false, 0},
		{"402", &APIError{HTTPStatus: 402}, false, 0},
		{"422", &APIError{HTTPStatus: 422}, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := retryDelayFor(tt.api, base)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.wantDur {
				t.Errorf("delay = %v, want %v", got, tt.wantDur)
			}
		})
	}
}
