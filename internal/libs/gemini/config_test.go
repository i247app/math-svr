package gemini

import (
	"errors"
	"testing"
	"time"
)

func validConfig() Config {
	return Config{
		APIKey: "test-key",
		Model:  "gemini-2.0-flash",
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"valid", func(c *Config) {}, false},
		{"valid with models/ prefix", func(c *Config) { c.Model = "models/gemini-2.0-flash" }, false},
		{"valid with base url override", func(c *Config) { c.BaseURL = "http://127.0.0.1:9999/v1beta" }, false},

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

// TestNormalizeModel: the model is part of the URL path, so an operator
// pasting either form out of the docs must not produce models/models/...
func TestNormalizeModel(t *testing.T) {
	tests := []struct{ in, want string }{
		{"gemini-2.0-flash", "models/gemini-2.0-flash"},
		{"models/gemini-2.0-flash", "models/gemini-2.0-flash"},
		{"  gemini-2.0-flash  ", "models/gemini-2.0-flash"},
		{"text-embedding-004", "models/text-embedding-004"},
	}
	for _, tt := range tests {
		if got := normalizeModel(tt.in); got != tt.want {
			t.Errorf("normalizeModel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// TestQuotaVersusThrottling: Gemini reuses 429 / RESOURCE_EXHAUSTED for a
// per-minute burst limit and for a daily quota. Only the first is worth
// retrying.
func TestQuotaVersusThrottling(t *testing.T) {
	burst := classifyAPIError(&APIError{
		HTTPStatus: 429,
		Status:     "RESOURCE_EXHAUSTED",
		Message:    "Resource has been exhausted (e.g. check quota).",
	})
	daily := classifyAPIError(&APIError{
		HTTPStatus: 429,
		Status:     "RESOURCE_EXHAUSTED",
		Message:    "You exceeded your current quota exceeded for the free tier",
	})

	if !errors.Is(burst, ErrRateLimited) {
		t.Errorf("burst 429 = %v, want ErrRateLimited", burst)
	}
	if IsConfigError(burst) {
		t.Error("burst 429 must not classify as a config error")
	}

	if !errors.Is(daily, ErrQuotaExhausted) {
		t.Errorf("daily 429 = %v, want ErrQuotaExhausted", daily)
	}
	if IsRateLimited(daily) {
		t.Error("daily quota must NOT classify as rate-limited — retrying it cannot succeed today")
	}
	if !IsConfigError(daily) {
		t.Error("daily quota must classify as a config error (needs an operator)")
	}
}

func TestIsAuthErrorByStatusString(t *testing.T) {
	// Gemini identifies these by the google.rpc status string as well as
	// the HTTP code, and proxies sometimes rewrite the latter.
	for _, st := range []string{"UNAUTHENTICATED", "PERMISSION_DENIED"} {
		err := classifyAPIError(&APIError{HTTPStatus: 200, Status: st})
		if !IsAuthError(err) {
			t.Errorf("status %s: IsAuthError() = false, want true", st)
		}
		if !IsConfigError(err) {
			t.Errorf("status %s: IsConfigError() = false, want true", st)
		}
	}
}

// TestFailedPreconditionIsConfigError: this is Gemini's "billing is not
// enabled" / "not available in your region" signal, which no amount of
// retrying fixes.
func TestFailedPreconditionIsConfigError(t *testing.T) {
	err := classifyAPIError(&APIError{HTTPStatus: 400, Status: "FAILED_PRECONDITION"})
	if !IsConfigError(err) {
		t.Error("FAILED_PRECONDITION must classify as a config error")
	}
}

func TestClassifyContextLength(t *testing.T) {
	err := classifyAPIError(&APIError{
		HTTPStatus: 400,
		Status:     "INVALID_ARGUMENT",
		Message:    "The input token count exceeds the maximum number of tokens allowed",
	})
	if !errors.Is(err, ErrContextTooLarge) {
		t.Errorf("error = %v, want ErrContextTooLarge", err)
	}

	plain := classifyAPIError(&APIError{
		HTTPStatus: 400,
		Status:     "INVALID_ARGUMENT",
		Message:    "contents is required",
	})
	if errors.Is(plain, ErrContextTooLarge) {
		t.Error("a plain 400 must not be classified as context-too-large")
	}
}

// TestCodeStringHandlesBothShapes: the docs describe a numeric google.rpc
// code and a snake_case one; both are observed.
func TestCodeStringHandlesBothShapes(t *testing.T) {
	if got := codeString(float64(429)); got != "429" {
		t.Errorf("numeric code = %q, want 429", got)
	}
	if got := codeString("rate_limit_exceeded"); got != "rate_limit_exceeded" {
		t.Errorf("string code = %q", got)
	}
	if got := codeString(nil); got != "" {
		t.Errorf("nil code = %q, want empty", got)
	}
}

// TestRetryInfoHint: Google puts the backoff hint inside error.details
// rather than in a Retry-After header.
func TestRetryInfoHint(t *testing.T) {
	w := &wireError{
		Code:    float64(429),
		Status:  "RESOURCE_EXHAUSTED",
		Message: "rate limited",
		Details: []struct {
			Type       string `json:"@type"`
			RetryDelay string `json:"retryDelay"`
		}{
			{Type: "type.googleapis.com/google.rpc.RetryInfo", RetryDelay: "17s"},
		},
	}
	if got := w.retryDelay(); got != 17*time.Second {
		t.Errorf("retryDelay() = %v, want 17s", got)
	}

	api := w.toAPIError(429, 0)
	if api.RetryAfter != 17*time.Second {
		t.Errorf("APIError.RetryAfter = %v, want 17s", api.RetryAfter)
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
		{"429 no hint", &APIError{HTTPStatus: 429}, false, 0},
		{"429 with hint", &APIError{HTTPStatus: 429, RetryAfter: 2 * time.Second}, true, 2 * time.Second},
		{"429 hint too long", &APIError{HTTPStatus: 429, RetryAfter: maxRetryAfter + time.Second}, false, 0},
		{
			"daily quota never retried",
			&APIError{HTTPStatus: 429, Message: "quota exceeded", RetryAfter: time.Second},
			false, 0,
		},
		{"500", &APIError{HTTPStatus: 500}, true, base},
		{"503", &APIError{HTTPStatus: 503}, true, base},
		{"403", &APIError{HTTPStatus: 403}, false, 0},
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
