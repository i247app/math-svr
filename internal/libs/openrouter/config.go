// Package openrouter is the math-svr libs wrapper around the OpenRouter
// REST API (https://openrouter.ai/api/v1/chat/completions). It owns the
// LLM client lifecycle (credentials, timeouts, retries) and exposes the
// same narrow chat surface that the adapter layer
// (`internal/adapter/bot`) consumes — mirroring the role of
// `internal/libs/langchain` and `internal/libs/eino`.
//
// Unlike those two, this package uses NO vendor SDK: every call goes out
// through `internal/shared/http_client`, the project's shared outbound
// HTTP client. There is no `http.Client` constructed here.
//
// This package never imports from `internal/domain`, `internal/application`,
// or `internal/adapter` — it is the lowest layer. Errors returned here are
// plain wrapped errors or typed sentinels (ErrInvalidConfig,
// ErrDecodeResponse, ErrContextTooLarge, ErrUnsupportedOp, ErrRateLimited);
// the adapter layer translates them into MathError(BOT_*) status codes.
//
// Secret handling: cfg.APIKey is read once at construction, set as the
// Authorization header on the shared client, and never logged from this
// package. The http_client default logging interceptor — which would dump
// both that header and the full prompt/response body — is deliberately
// disabled in NewClient; see the comment there.
package openrouter

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
var ErrInvalidConfig = errors.New("openrouter: invalid config")

// DefaultBaseURL is OpenRouter's v1 REST root. Config.BaseURL overrides it
// (used by tests to point at an httptest server).
const DefaultBaseURL = "https://openrouter.ai/api/v1"

// chatCompletionsPath is appended to BaseURL for every chat call.
const chatCompletionsPath = "/chat/completions"

// Config holds parameters for the OpenRouter-backed Client.
//
// There is no Backend field: OpenRouter is itself the router, and the
// vendor is selected by the Model id ("vendor/model-name", e.g.
// "openai/gpt-4o-mini", "anthropic/claude-3.5-sonnet").
//
// APIKey is the long-lived vendor credential and is sensitive — see
// package doc.
type Config struct {
	// APIKey is the OpenRouter credential (env BOT_OPENROUTER_API_KEY).
	// Required. Sent as "Authorization: Bearer <key>". NEVER log.
	APIKey string

	// BaseURL overrides DefaultBaseURL. Optional.
	BaseURL string

	// Model is the default chat model id in OpenRouter's
	// "vendor/model-name" form. Required — there is deliberately no
	// built-in default, because picking one silently would pick a price
	// and a vendor on the operator's behalf. Falls back at call time when
	// ChatRequest.Model is empty.
	Model string

	// SiteURL is sent as the optional "HTTP-Referer" header. Attribution
	// only: it governs whether the app appears on OpenRouter's rankings
	// and has no effect on whether a call succeeds.
	SiteURL string

	// AppTitle is sent as the optional "X-OpenRouter-Title" header. Same
	// attribution-only role as SiteURL.
	AppTitle string

	// Temperature is the default sampling temperature. A negative value
	// means "do not send any temperature override" so the vendor default
	// applies.
	Temperature float64

	// TopP is the default nucleus sampling cutoff. A negative value means
	// "do not send any top_p override" so the vendor default applies.
	TopP float64

	// MaxTokens caps response length. Zero means "do not send any limit"
	// so the vendor default applies.
	MaxTokens int

	// Timeout bounds a single LLM call. applyDefaults sets 60s when zero.
	Timeout time.Duration

	// MaxRetries is the per-call retry budget for transient failures.
	// applyDefaults sets 2 when zero. Same policy as libs/eino: only
	// transport errors, HTTP 408/502/503/504 and Retry-After-carrying 429s
	// are re-issued.
	MaxRetries int

	// RetryDelay is the base linear sleep between retries. applyDefaults
	// sets 500ms when zero.
	RetryDelay time.Duration

	// RequireAtBoot governs NewClient's startup probe. true → fail fast on
	// a bad credential. false → log and continue so dev environments
	// without LLM credentials still boot.
	RequireAtBoot bool
}

const (
	DefaultTimeout    = 60 * time.Second
	DefaultMaxRetries = 2
	DefaultRetryDelay = 500 * time.Millisecond

	// maxRetryAfter caps how long a Retry-After header may park a request.
	// A 429 asking for longer than this is surfaced to the caller instead
	// of being slept through, so one throttled call cannot hold an HTTP
	// handler open for minutes.
	maxRetryAfter = 30 * time.Second
)

// Validate enforces the credential invariants and returns errors wrapped
// with ErrInvalidConfig. Deliberately the same shape of rules as
// libs/eino.Config.Validate so the providers misconfigure the same way.
func (c Config) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return fmt.Errorf("%w: APIKey is required", ErrInvalidConfig)
	}
	if strings.TrimSpace(c.Model) == "" {
		return fmt.Errorf("%w: Model is required (e.g. \"openai/gpt-4o-mini\")", ErrInvalidConfig)
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
	if strings.TrimSpace(c.BaseURL) == "" {
		c.BaseURL = DefaultBaseURL
	}
	c.BaseURL = strings.TrimSuffix(c.BaseURL, "/")
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
