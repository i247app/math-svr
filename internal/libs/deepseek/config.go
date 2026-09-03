// Package deepseek is the math-svr libs wrapper around the DeepSeek API
// (https://api.deepseek.com). It owns the LLM client lifecycle
// (credentials, timeouts, retries) and exposes the same narrow chat
// surface the adapter layer (`internal/adapter/bot`) consumes — mirroring
// libs/langchain, libs/eino, libs/openrouter, libs/openai and libs/gemini.
//
// Like the other three direct clients, this package uses NO vendor SDK:
// every call goes out through `internal/shared/http_client`. There is no
// `http.Client` constructed here.
//
// # Why a separate package from libs/openai
//
// DeepSeek is deliberately OpenAI-compatible, so this is the closest
// sibling of the five and the question was real. It diverges on four
// points, and the first two are not cosmetic:
//
//   - Token cap. DeepSeek takes `max_tokens`. libs/openai always sends
//     `max_completion_tokens`, because OpenAI's reasoning models reject
//     the old name. Pointing libs/openai at api.deepseek.com would send a
//     field DeepSeek does not read, and the cap would silently not apply.
//   - Billing. DeepSeek signals an empty balance with HTTP 402, the way
//     OpenRouter does. libs/openai has no 402 path at all — it discovers
//     billing through a 429 carrying credit_balance_exhausted.
//   - Thinking mode. The `thinking` request object and the
//     `reasoning_content` / `reasoning_tokens` response fields have no
//     OpenAI equivalent.
//   - Capabilities. There is no embeddings endpoint, so Embed is
//     unsupported here while libs/openai implements it.
//
// Making libs/openai serve both would mean a vendor flag switching the
// token-cap field name, a second billing path, an optional thinking
// block and a capability toggle — four seams for two implementations.
// The repo's convention is one libs package per provider; this follows it.
// See the PR notes for the measured cost of that choice.
//
// This package never imports from `internal/domain`, `internal/application`,
// or `internal/adapter` — it is the lowest layer. Errors returned here are
// plain wrapped errors or typed sentinels; the adapter layer translates
// them into MathError(BOT_*) status codes.
//
// Secret handling: cfg.APIKey is read once at construction, set as the
// Authorization header, and never logged from this package. The
// http_client default logging interceptor — which would dump both that
// header and the full prompt/response body — is deliberately disabled in
// NewClient; see the comment there.
package deepseek

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
var ErrInvalidConfig = errors.New("deepseek: invalid config")

// DefaultBaseURL is DeepSeek's REST root. The vendor also serves the same
// API under /v1 for drop-in OpenAI SDK use; that suffix carries no version
// meaning and is not needed here. Config.BaseURL overrides it (used by
// tests to point at an httptest server, and to reach the /beta surface).
const DefaultBaseURL = "https://api.deepseek.com"

const chatCompletionsPath = "/chat/completions"

// Thinking-mode values accepted by Config.Thinking and
// Config.ReasoningEffort. Empty means "send nothing" so the model's own
// default stands.
const (
	ThinkingEnabled  = "enabled"
	ThinkingDisabled = "disabled"

	ReasoningEffortLow  = "low"
	ReasoningEffortHigh = "high"
	ReasoningEffortMax  = "max"
)

// Config holds parameters for the DeepSeek-backed Client.
//
// APIKey is the long-lived vendor credential and is sensitive — see
// package doc.
type Config struct {
	// APIKey is the DeepSeek credential (env BOT_DEEPSEEK_API_KEY).
	// Required. Sent as "Authorization: Bearer <key>". NEVER log.
	APIKey string

	// BaseURL overrides DefaultBaseURL. Optional. Point it at
	// https://api.deepseek.com/beta to reach the beta surface.
	BaseURL string

	// Model is the default model id (e.g. "deepseek-v4-flash"). Required —
	// there is deliberately no built-in default, because picking one
	// silently would pick a price on the operator's behalf.
	Model string

	// Thinking turns the reasoning block on or off explicitly
	// ("enabled" / "disabled"). Empty sends nothing, leaving the model's
	// own default in place.
	//
	// Worth setting: on a thinking-capable model the reasoning tokens are
	// billed on top of the answer, so a deploy that only needs quiz JSON
	// can turn it off and stop paying for deliberation it discards.
	Thinking string

	// ReasoningEffort tunes how much deliberation to spend
	// ("low" / "high" / "max"). Empty sends nothing. Only meaningful
	// alongside Thinking enabled (or a model that thinks by default).
	ReasoningEffort string

	// Temperature is the default sampling temperature. A negative value
	// means "do not send any temperature override" so the model default
	// applies.
	Temperature float64

	// TopP is the default nucleus sampling cutoff. A negative value means
	// "do not send any top_p override" so the model default applies.
	TopP float64

	// MaxTokens caps response length. Zero means "do not send any limit".
	// Sent as `max_tokens` — see the package doc on why that differs from
	// libs/openai.
	MaxTokens int

	// Timeout bounds a single call. applyDefaults sets 60s when zero.
	//
	// Worth raising for thinking-mode models: deliberation happens before
	// the first token, so a reasoning call can sit silent far longer than
	// a plain completion.
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
)

// Validate enforces the credential and enum invariants and returns errors
// wrapped with ErrInvalidConfig. Same shape of rules as the sibling libs
// packages so every provider misconfigures the same way.
func (c Config) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return fmt.Errorf("%w: APIKey is required", ErrInvalidConfig)
	}
	if strings.TrimSpace(c.Model) == "" {
		return fmt.Errorf("%w: Model is required (e.g. \"deepseek-v4-flash\")", ErrInvalidConfig)
	}
	if c.BaseURL != "" &&
		!strings.HasPrefix(c.BaseURL, "http://") &&
		!strings.HasPrefix(c.BaseURL, "https://") {
		return fmt.Errorf("%w: BaseURL must use http(s):// (got %q)", ErrInvalidConfig, c.BaseURL)
	}
	// Catch a typo at boot rather than on the first quiz of the day: an
	// unknown enum here is rejected upstream as a 400 on every call.
	switch c.Thinking {
	case "", ThinkingEnabled, ThinkingDisabled:
	default:
		return fmt.Errorf("%w: Thinking must be %q or %q (got %q)",
			ErrInvalidConfig, ThinkingEnabled, ThinkingDisabled, c.Thinking)
	}
	switch c.ReasoningEffort {
	case "", ReasoningEffortLow, ReasoningEffortHigh, ReasoningEffortMax:
	default:
		return fmt.Errorf("%w: ReasoningEffort must be %q, %q or %q (got %q)",
			ErrInvalidConfig, ReasoningEffortLow, ReasoningEffortHigh,
			ReasoningEffortMax, c.ReasoningEffort)
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
