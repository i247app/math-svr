package openai

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validConfig() Config {
	return Config{
		APIKey: "test-key",
		Model:  "gpt-4.1",
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"valid", func(c *Config) {}, false},
		{"valid with base url override", func(c *Config) { c.BaseURL = "http://127.0.0.1:9999/v1" }, false},
		{"valid with metadata", func(c *Config) {
			c.Store = true
			c.Metadata = map[string]string{"env": "production"}
		}, false},

		{"missing api key", func(c *Config) { c.APIKey = "  " }, true},
		{"missing model", func(c *Config) { c.Model = "" }, true},
		{"bad base url scheme", func(c *Config) { c.BaseURL = "ftp://x" }, true},
		{"too many metadata pairs", func(c *Config) {
			c.Metadata = map[string]string{}
			for i := range maxMetadataPairs + 1 {
				c.Metadata[string(rune('a'+i))] = "v"
			}
		}, true},
		{"metadata key too long", func(c *Config) {
			c.Metadata = map[string]string{strings.Repeat("k", maxMetadataKeyBytes+1): "v"}
		}, true},
		{"metadata value too long", func(c *Config) {
			c.Metadata = map[string]string{"k": strings.Repeat("v", maxMetadataValueBytes+1)}
		}, true},
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

// TestQuotaVersusThrottling is the asymmetry that separates this client
// from libs/openrouter: OpenAI reuses HTTP 429 for an exhausted credit
// balance, which must NOT be treated as transient throttling.
func TestQuotaVersusThrottling(t *testing.T) {
	throttle := classifyAPIError(&APIError{HTTPStatus: 429, Type: "rate_limit_error"})
	quota := classifyAPIError(&APIError{
		HTTPStatus: 429,
		Type:       "rate_limit_error",
		Code:       "credit_balance_exhausted",
	})

	if !errors.Is(throttle, ErrRateLimited) {
		t.Errorf("plain 429 = %v, want ErrRateLimited", throttle)
	}
	if IsConfigError(throttle) {
		t.Error("plain 429 must not classify as a config error")
	}

	if !errors.Is(quota, ErrQuotaExhausted) {
		t.Errorf("billing 429 = %v, want ErrQuotaExhausted", quota)
	}
	if IsRateLimited(quota) {
		t.Error("billing 429 must NOT classify as rate-limited — retrying it can never succeed")
	}
	if !IsConfigError(quota) {
		t.Error("billing 429 must classify as a config error (needs an operator)")
	}

	// The legacy code name is still observed in the wild.
	legacy := classifyAPIError(&APIError{HTTPStatus: 429, Code: "insufficient_quota"})
	if !errors.Is(legacy, ErrQuotaExhausted) {
		t.Errorf("insufficient_quota = %v, want ErrQuotaExhausted", legacy)
	}
}

func TestClassifyContextLength(t *testing.T) {
	byCode := classifyAPIError(&APIError{HTTPStatus: 400, Code: "context_length_exceeded"})
	if !errors.Is(byCode, ErrContextTooLarge) {
		t.Errorf("by code = %v, want ErrContextTooLarge", byCode)
	}

	byMessage := classifyAPIError(&APIError{
		HTTPStatus: 400,
		Message:    "This model's maximum context length is 128000 tokens",
	})
	if !errors.Is(byMessage, ErrContextTooLarge) {
		t.Errorf("by message = %v, want ErrContextTooLarge", byMessage)
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
		{
			"billing 429 never retried",
			&APIError{HTTPStatus: 429, Code: "credit_balance_exhausted", RetryAfter: time.Second},
			false, 0,
		},
		{"500", &APIError{HTTPStatus: 500}, true, base},
		{"503", &APIError{HTTPStatus: 503}, true, base},
		{"401", &APIError{HTTPStatus: 401}, false, 0},
		{"400", &APIError{HTTPStatus: 400}, false, 0},
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

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		in   string
		want time.Duration
	}{
		{"", 0},
		{"3", 3 * time.Second},
		{"0", 0},
		{"-5", 0},
		{"not-a-number", 0},
	}
	for _, tt := range tests {
		if got := parseRetryAfter(tt.in); got != tt.want {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
