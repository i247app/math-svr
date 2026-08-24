// Package eino is the math-svr libs wrapper around cloudwego/eino (+ the
// eino-ext model components). It owns the LLM client lifecycle (vendor
// selection, timeouts, retries) and exposes a narrow chat surface that the
// adapter layer (`internal/adapter/bot`) consumes — mirroring the role of
// `internal/libs/langchain` for the langchaingo-backed provider.
//
// This package never imports from `internal/domain`, `internal/application`,
// or `internal/adapter` — it is the lowest layer. Errors returned here are
// plain wrapped errors or typed sentinels (ErrInvalidConfig,
// ErrDecodeResponse, ErrContextTooLarge, ErrUnsupportedOp, ErrRateLimited);
// the adapter layer translates them into MathError(BOT_*) status codes.
//
// Secret handling: cfg.APIKey is read once at construction, passed into the
// eino-ext vendor constructor, and never logged from this package.
package eino

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidConfig is the sentinel wrapped by Config.Validate. The adapter
// factory uses errors.Is to distinguish "config is wrong"
// (→ BOT_CONFIG_INVALID) from "config is fine but the network or upstream
// said no" (→ BOT_CONNECT_FAILED).
var ErrInvalidConfig = errors.New("eino: invalid config")

// Backend identifies which LLM vendor the eino Client will be configured
// against. The names intentionally match libs/langchain so a deploy can
// switch providers without relearning vendor vocabulary; the mapping onto
// eino-ext component packages (googleai→gemini, anthropic→claude) is an
// implementation detail of newModels.
type Backend string

const (
	BackendGoogleAI  Backend = "googleai"
	BackendOpenAI    Backend = "openai"
	BackendAnthropic Backend = "anthropic"
	BackendOllama    Backend = "ollama"
)

// Config holds parameters for the eino-backed Client.
//
// APIKey is the long-lived vendor credential and is sensitive — see
// package doc.
type Config struct {
	// Backend selects which LLM vendor to instantiate.
	Backend Backend

	// APIKey is the vendor credential. Required for googleai / openai /
	// anthropic; ignored for ollama (which uses BaseURL only). NEVER log.
	APIKey string

	// BaseURL overrides the vendor's default REST endpoint. Empty resolves
	// to the vendor default. Required for ollama (e.g. "http://localhost:11434").
	BaseURL string

	// Model is the default chat model id (e.g. "gemini-1.5-flash",
	// "gpt-4o-mini"). Required; falls back at call time when
	// ChatRequest.Model is empty.
	Model string

	// Temperature is the default sampling temperature. A negative value means
	// "do not pass any temperature override" so the vendor default applies.
	Temperature float64

	// TopP is the default nucleus sampling cutoff. A negative value means
	// "do not pass any topP override" so the vendor default applies.
	TopP float64

	// MaxTokens caps response length. Zero means "do not pass any limit"
	// so the vendor default applies (except anthropic, where the API
	// mandates max_tokens — see defaultAnthropicMaxTokens).
	MaxTokens int

	// Timeout bounds a single LLM call. applyDefaults sets 60s when zero.
	Timeout time.Duration

	// MaxRetries is the per-call retry budget for transient failures.
	// applyDefaults sets 2 when zero. Enforced inside Client by re-issuing
	// on classified-as-retryable errors only, same policy as libs/langchain.
	MaxRetries int

	// RetryDelay is the base linear sleep between retries. applyDefaults
	// sets 500ms when zero.
	RetryDelay time.Duration

	// Store asks the upstream to persist the prompt and the completion so
	// they show up on the vendor's dashboard log page. OPENAI BACKEND ONLY
	// — it maps to the chat-completions `store: true` field, which defaults
	// to false, and is what makes a request visible at
	// platform.openai.com/logs. Silently ignored on googleai / anthropic /
	// ollama, which have no equivalent.
	//
	// Opt-in on purpose: turning it on ships prompt + response content to
	// OpenAI's storage (30 days, per their data-controls guide) and makes
	// it readable by anyone with dashboard access. Quiz prompts carry
	// curriculum context and student answers.
	Store bool

	// RequireAtBoot governs NewClient's startup probe. true → fail fast
	// on a bad credential. false → log a warning and continue so dev
	// environments without LLM credentials still boot.
	RequireAtBoot bool
}

const (
	DefaultTimeout    = 60 * time.Second
	DefaultMaxRetries = 2
	DefaultRetryDelay = 500 * time.Millisecond

	// defaultAnthropicMaxTokens is applied when Config.MaxTokens is zero on
	// the anthropic backend: the Anthropic Messages API requires max_tokens
	// on every request, so "vendor default" does not exist there.
	defaultAnthropicMaxTokens = 4096
)

// Validate enforces the credential / backend invariants and returns errors
// wrapped with ErrInvalidConfig. Deliberately identical rules to
// libs/langchain.Config.Validate so the two providers misconfigure the
// same way.
func (c Config) Validate() error {
	switch c.Backend {
	case BackendGoogleAI, BackendOpenAI, BackendAnthropic:
		if strings.TrimSpace(c.APIKey) == "" {
			return fmt.Errorf("%w: APIKey is required for backend %q", ErrInvalidConfig, c.Backend)
		}
	case BackendOllama:
		if strings.TrimSpace(c.BaseURL) == "" {
			return fmt.Errorf("%w: BaseURL is required for ollama backend", ErrInvalidConfig)
		}
	case "":
		return fmt.Errorf("%w: Backend is required", ErrInvalidConfig)
	default:
		return fmt.Errorf("%w: unsupported backend %q", ErrInvalidConfig, c.Backend)
	}
	if strings.TrimSpace(c.Model) == "" {
		return fmt.Errorf("%w: Model is required", ErrInvalidConfig)
	}
	if c.BaseURL != "" &&
		!strings.HasPrefix(c.BaseURL, "http://") &&
		!strings.HasPrefix(c.BaseURL, "https://") {
		return fmt.Errorf("%w: BaseURL must use http(s):// (got %q)", ErrInvalidConfig, c.BaseURL)
	}
	if c.Timeout < 0 {
		return fmt.Errorf("%w: Timeout must be >= 0", ErrInvalidConfig)
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("%w: MaxRetries must be >= 0", ErrInvalidConfig)
	}
	if c.RetryDelay < 0 {
		return fmt.Errorf("%w: RetryDelay must be >= 0", ErrInvalidConfig)
	}
	return nil
}

// applyDefaults substitutes zero values with sensible defaults. Called
// after Validate so callers can leave the transport-tuning fields blank.
func (c *Config) applyDefaults() {
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = DefaultMaxRetries
	}
	if c.RetryDelay == 0 {
		c.RetryDelay = DefaultRetryDelay
	}
}
