// Package openai is the math-svr libs wrapper around OpenAI's REST API
// (https://api.openai.com/v1). It owns the LLM client lifecycle
// (credentials, timeouts, retries) and exposes the same narrow chat
// surface the adapter layer (`internal/adapter/bot`) consumes — mirroring
// the role of `internal/libs/langchain`, `internal/libs/eino` and
// `internal/libs/openrouter`.
//
// Like libs/openrouter and unlike the other two, this package uses NO
// vendor SDK: every call goes out through `internal/shared/http_client`,
// the project's shared outbound HTTP client. There is no `http.Client`
// constructed here.
//
// # Why a separate package from libs/openrouter
//
// OpenRouter speaks the OpenAI chat-completions schema, so the two
// clients look alike at a glance. They diverge on everything that
// matters operationally:
//
//   - Error envelope. OpenAI returns {"error":{message,type,param,code}}
//     with a STRING code ("credit_balance_exhausted"); OpenRouter returns
//     a NUMERIC code equal to the HTTP status.
//   - Billing exhaustion. OpenRouter signals it with HTTP 402; OpenAI
//     reuses 429 with type=rate_limit_error + code=credit_balance_exhausted,
//     which must NOT be retried even though a plain 429 should be.
//   - Embeddings. OpenAI has /v1/embeddings, so Embed is genuinely
//     supported here. OpenRouter has no embedding endpoint.
//   - Dashboard logging. `store` + `metadata` are OpenAI-only and are the
//     reason this provider exists (platform.openai.com/logs).
//   - max_tokens is deprecated upstream and rejected outright by the
//     reasoning models; this client always sends max_completion_tokens.
//   - Headers. OpenAI-Organization / OpenAI-Project vs OpenRouter's
//     HTTP-Referer / X-OpenRouter-Title attribution pair.
//
// Folding both into one parametrised client would need a pluggable error
// parser, a pluggable header set, an optional embeddings path and two
// different retry rules for the same status code — a seam per difference,
// with two implementations. The repo's own convention is one libs package
// per provider, so this follows it. See the note in the PR description
// about extracting a shared SSE/retry core if a third REST provider lands.
//
// # API choice
//
// This client targets Chat Completions, not the newer Responses API. The
// bot adapter's ChatRequest is a messages/role/content shape that maps 1:1
// onto chat-completions, and both existing OpenAI paths (the eino openai
// backend and openrouter) already speak it — so prompts, JSON mode and
// results stay directly comparable across all four providers, which is the
// entire point of keeping them switchable. Responses' advantages (built-in
// tools, server-side state, reasoning items) are unused by quiz and
// exercise generation, and its `store` defaults to true, which would flip
// a data-retention decision silently.
//
// This package never imports from `internal/domain`, `internal/application`,
// or `internal/adapter` — it is the lowest layer. Errors returned here are
// plain wrapped errors or typed sentinels; the adapter layer translates
// them into MathError(BOT_*) status codes.
//
// Secret handling: cfg.APIKey is read once at construction, set as the
// Authorization header on the shared client, and never logged from this
// package. The http_client default logging interceptor — which would dump
// both that header and the full prompt/response body — is deliberately
// disabled in NewClient; see the comment there.
package openai

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
var ErrInvalidConfig = errors.New("openai: invalid config")

// DefaultBaseURL is OpenAI's v1 REST root. Config.BaseURL overrides it
// (used by tests to point at an httptest server).
const DefaultBaseURL = "https://api.openai.com/v1"

const (
	chatCompletionsPath = "/chat/completions"
	embeddingsPath      = "/embeddings"
)

// Config holds parameters for the OpenAI-backed Client.
//
// There is no Backend field: this client talks to OpenAI and only OpenAI.
// The model id is a bare OpenAI name ("gpt-4.1"), NOT the "vendor/model"
// form OpenRouter uses.
//
// APIKey is the long-lived vendor credential and is sensitive — see
// package doc.
type Config struct {
	// APIKey is the OpenAI credential (env BOT_OPENAI_API_KEY). Required.
	// Sent as "Authorization: Bearer <key>". NEVER log.
	APIKey string

	// BaseURL overrides DefaultBaseURL. Optional.
	BaseURL string

	// Model is the default chat model id (e.g. "gpt-4.1"). Required —
	// there is deliberately no built-in default, because picking one
	// silently would pick a price on the operator's behalf. Falls back at
	// call time when ChatRequest.Model is empty.
	Model string

	// EmbedModel is the default embedding model id (e.g.
	// "text-embedding-3-small"). Optional; Embed falls back to
	// EmbedRequest.Model, then EmbedModel, then Model — same precedence as
	// libs/langchain.
	EmbedModel string

	// Organization sets the "OpenAI-Organization" header. Optional; only
	// needed on accounts belonging to more than one org.
	Organization string

	// Project sets the "OpenAI-Project" header. Optional. Useful when a
	// legacy user-level key must be attributed to a specific project;
	// project-scoped keys (sk-proj-...) already carry it.
	Project string

	// Store sends `store: true` so the request and response are retained
	// by OpenAI and become visible at platform.openai.com/logs.
	//
	// Opt-in on purpose: it ships prompt + response content to OpenAI's
	// storage (30 days) readable by anyone with dashboard access. Quiz
	// prompts carry curriculum context and student answers. Organisations
	// with Zero Data Retention have `store` forced to false upstream
	// regardless of what is sent here.
	Store bool

	// Metadata is attached to every stored completion so the dashboard can
	// be filtered (e.g. {"env":"production","feature":"quiz"}). Only
	// meaningful alongside Store. OpenAI caps this at 16 pairs, 64-char
	// keys and 512-char values; Validate enforces those limits.
	Metadata map[string]string

	// Temperature is the default sampling temperature. A negative value
	// means "do not send any temperature override" so the model default
	// applies.
	Temperature float64

	// TopP is the default nucleus sampling cutoff. A negative value means
	// "do not send any top_p override" so the model default applies.
	TopP float64

	// MaxTokens caps response length. Zero means "do not send any limit".
	// Sent as `max_completion_tokens` — see buildRequest.
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

	// maxRetryAfter caps how long a Retry-After header may park a request,
	// so one throttled call cannot hold an HTTP handler open for minutes.
	maxRetryAfter = 30 * time.Second

	// OpenAI's documented metadata limits.
	maxMetadataPairs      = 16
	maxMetadataKeyBytes   = 64
	maxMetadataValueBytes = 512
)

// Validate enforces the credential invariants and returns errors wrapped
// with ErrInvalidConfig. Same shape of rules as the sibling libs packages
// so every provider misconfigures the same way.
func (c Config) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return fmt.Errorf("%w: APIKey is required", ErrInvalidConfig)
	}
	if strings.TrimSpace(c.Model) == "" {
		return fmt.Errorf("%w: Model is required (e.g. \"gpt-4.1\")", ErrInvalidConfig)
	}
	if c.BaseURL != "" &&
		!strings.HasPrefix(c.BaseURL, "http://") &&
		!strings.HasPrefix(c.BaseURL, "https://") {
		return fmt.Errorf("%w: BaseURL must use http(s):// (got %q)", ErrInvalidConfig, c.BaseURL)
	}
	// Reject an oversized metadata map here rather than letting OpenAI
	// reject the whole completion at request time.
	if len(c.Metadata) > maxMetadataPairs {
		return fmt.Errorf("%w: Metadata has %d pairs, OpenAI allows %d",
			ErrInvalidConfig, len(c.Metadata), maxMetadataPairs)
	}
	for k, v := range c.Metadata {
		if len(k) > maxMetadataKeyBytes {
			return fmt.Errorf("%w: Metadata key %q exceeds %d bytes",
				ErrInvalidConfig, k, maxMetadataKeyBytes)
		}
		if len(v) > maxMetadataValueBytes {
			return fmt.Errorf("%w: Metadata value for key %q exceeds %d bytes",
				ErrInvalidConfig, k, maxMetadataValueBytes)
		}
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
