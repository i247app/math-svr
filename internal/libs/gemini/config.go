// Package gemini is the math-svr libs wrapper around Google's Gemini API
// (https://generativelanguage.googleapis.com/v1beta). It owns the LLM
// client lifecycle (credentials, timeouts, retries) and exposes the same
// narrow chat surface the adapter layer (`internal/adapter/bot`) consumes
// — mirroring libs/langchain, libs/eino, libs/openrouter and libs/openai.
//
// Like libs/openrouter and libs/openai, this package uses NO vendor SDK:
// every call goes out through `internal/shared/http_client`. There is no
// `http.Client` constructed here.
//
// # Why this is not a variant of libs/openai
//
// The two OpenAI-compatible clients in this repo differ in their details.
// Gemini differs in its SHAPE, so almost nothing is shareable:
//
//   - The model is part of the URL PATH and the operation is a suffix on
//     it (models/{model}:generateContent), not a field in the body.
//   - Roles are "user" and "model" only. There is no system role — system
//     prompts move to a separate top-level systemInstruction object.
//   - Message text is nested as contents[].parts[].text, not a flat
//     content string, and sampling options live under generationConfig
//     with different names (maxOutputTokens, stopSequences).
//   - The reply is candidates[].content.parts[], with token counts under
//     usageMetadata (promptTokenCount / candidatesTokenCount).
//   - Auth is the x-goog-api-key header, not a bearer token.
//   - Safety filtering is a first-class outcome with no OpenAI analogue: a
//     request can succeed with HTTP 200 and return zero candidates because
//     the prompt was blocked.
//   - Errors are google.rpc shaped, with a status string and an optional
//     RetryInfo in details[].
//
// # API choice
//
// This client targets the v1beta REST surface, which is what every current
// Gemini model is served from and what the official docs document. The
// bot adapter's ChatRequest is translated in buildRequest; see the role
// mapping there, which is the only genuinely lossy step.
//
// This package never imports from `internal/domain`, `internal/application`,
// or `internal/adapter` — it is the lowest layer. Errors returned here are
// plain wrapped errors or typed sentinels; the adapter layer translates
// them into MathError(BOT_*) status codes.
//
// Secret handling: cfg.APIKey is read once at construction and set as the
// x-goog-api-key header. It is never placed in the URL — Gemini also
// accepts ?key=<secret>, which would put the credential into every access
// log, proxy trace and error message. It is never logged from this
// package, and the http_client default logging interceptor is disabled in
// NewClient; see the comment there.
package gemini

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
var ErrInvalidConfig = errors.New("gemini: invalid config")

// DefaultBaseURL is the Gemini v1beta REST root. Config.BaseURL overrides
// it (used by tests to point at an httptest server).
const DefaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// apiKeyHeader carries the credential. Deliberately the header form: the
// ?key= query-parameter alternative would leak the secret into access
// logs and proxy traces.
const apiKeyHeader = "x-goog-api-key"

// Config holds parameters for the Gemini-backed Client.
//
// There is no Backend field: this client talks to Gemini and only Gemini.
//
// APIKey is the long-lived vendor credential and is sensitive — see
// package doc.
type Config struct {
	// APIKey is the Gemini credential (env BOT_GEMINI_API_KEY). Required.
	// Sent as the x-goog-api-key header. NEVER log.
	APIKey string

	// BaseURL overrides DefaultBaseURL. Optional.
	BaseURL string

	// Model is the default model id (e.g. "gemini-2.0-flash"). Required —
	// there is deliberately no built-in default, because picking one
	// silently would pick a price on the operator's behalf. A "models/"
	// prefix is optional; normalizeModel adds or keeps exactly one.
	Model string

	// EmbedModel is the default embedding model id (e.g.
	// "text-embedding-004"). Optional; Embed falls back to
	// EmbedRequest.Model, then EmbedModel, then Model — same precedence as
	// libs/langchain and libs/openai.
	EmbedModel string

	// Temperature is the default sampling temperature. A negative value
	// means "do not send any temperature override" so the model default
	// applies.
	Temperature float64

	// TopP is the default nucleus sampling cutoff. A negative value means
	// "do not send any topP override" so the model default applies.
	TopP float64

	// MaxTokens caps response length. Zero means "do not send any limit".
	// Sent as generationConfig.maxOutputTokens.
	MaxTokens int

	// Timeout bounds a single call. applyDefaults sets 60s when zero.
	Timeout time.Duration

	// MaxRetries is the per-call retry budget for transient failures.
	// applyDefaults sets 2 when zero.
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

	// maxRetryAfter caps how long a RetryInfo hint may park a request, so
	// one throttled call cannot hold an HTTP handler open for minutes.
	maxRetryAfter = 30 * time.Second
)

// Validate enforces the credential invariants and returns errors wrapped
// with ErrInvalidConfig. Same shape of rules as the sibling libs packages
// so every provider misconfigures the same way.
func (c Config) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return fmt.Errorf("%w: APIKey is required", ErrInvalidConfig)
	}
	if strings.TrimSpace(c.Model) == "" {
		return fmt.Errorf("%w: Model is required (e.g. \"gemini-2.0-flash\")", ErrInvalidConfig)
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

// normalizeModel renders a model id as the "models/<id>" form the URL path
// requires, accepting either form from config so an operator copying an id
// straight out of the docs cannot produce "models/models/gemini-2.0-flash".
func normalizeModel(model string) string {
	m := strings.TrimSpace(model)
	m = strings.TrimPrefix(m, "models/")
	return "models/" + m
}
