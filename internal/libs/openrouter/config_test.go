package openrouter

import (
	"errors"
	"testing"
	"time"
)

func validConfig() Config {
	return Config{
		APIKey: "test-key",
		Model:  "openai/gpt-4o-mini",
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"valid", func(c *Config) {}, false},
		{"valid with base url override", func(c *Config) { c.BaseURL = "http://127.0.0.1:9999/api/v1" }, false},

		{"missing api key", func(c *Config) { c.APIKey = "  " }, true},
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

func TestConfigApplyDefaultsTrimsBaseURLSlash(t *testing.T) {
	cfg := validConfig()
	cfg.BaseURL = "http://127.0.0.1:9999/api/v1/"
	cfg.applyDefaults()

	if cfg.BaseURL != "http://127.0.0.1:9999/api/v1" {
		t.Errorf("BaseURL = %q, want the trailing slash trimmed", cfg.BaseURL)
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		in   string
		want time.Duration
	}{
		{"", 0},
		{"  ", 0},
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

func TestCodeString(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, ""},
		{"string", "insufficient_quota", "insufficient_quota"},
		{"whole number", float64(429), "429"},
		{"fractional", 1.5, "1.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := codeString(tt.in); got != tt.want {
				t.Errorf("codeString(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
