package openrouter

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Sentinel errors surfaced by the Client. Each maps to a distinct
// MathError(BOT_*) code in the adapter layer; see adapter/bot/errors.go.
// The taxonomy is identical to libs/langchain and libs/eino so
// mapOpenRouterError stays symmetrical with its two siblings.
var (
	// ErrDecodeResponse signals that a call returned 2xx but the body the
	// client expected was missing or unparsable. Adapter maps to
	// BOT_SERIALIZE_FAILED.
	ErrDecodeResponse = errors.New("openrouter: failed to decode response")

	// ErrContextTooLarge signals that the input exceeded the model's
	// context window. Adapter maps to BOT_CONTEXT_TOO_LARGE.
	ErrContextTooLarge = errors.New("openrouter: context window exceeded")

	// ErrUnsupportedOp signals that OpenRouter does not support the
	// requested capability. Adapter maps to BOT_UNSUPPORTED_OP.
	ErrUnsupportedOp = errors.New("openrouter: operation not supported")

	// ErrRateLimited signals a 429. Adapter maps to BOT_RATE_LIMITED.
	ErrRateLimited = errors.New("openrouter: rate limit exceeded")

	// ErrInsufficientCredits signals a 402 — the account is out of credit.
	// Non-recoverable within the call and needs operator action, so the
	// adapter maps it to BOT_CONFIG_INVALID alongside the auth failures
	// rather than to a transient code.
	ErrInsufficientCredits = errors.New("openrouter: insufficient credits")
)

// APIError is the typed shape of an OpenRouter error response. The client
// lifts every non-2xx body (and every 200-with-error-envelope body) into
// this shape in translateError.
//
// Message carries the vendor's own error text — never the request prompt
// and never the raw response body, so it is safe to log and to surface in
// the MathError debug field.
type APIError struct {
	HTTPStatus int
	Code       string
	Message    string
	// RetryAfter is the parsed `Retry-After` header (seconds) on a 429.
	// Zero when absent.
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e == nil {
		return "openrouter: <nil APIError>"
	}
	if e.Code != "" {
		return fmt.Sprintf("openrouter: http %d code %s: %s", e.HTTPStatus, e.Code, e.Message)
	}
	if e.HTTPStatus != 0 {
		return fmt.Sprintf("openrouter: http %d: %s", e.HTTPStatus, e.Message)
	}
	return "openrouter: " + e.Message
}

// IsAuthError reports whether err is an authentication / authorization
// failure (401 or 403 — the latter is also OpenRouter's moderation block).
func IsAuthError(err error) bool {
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	return api.HTTPStatus == 401 || api.HTTPStatus == 403
}

// IsRateLimited reports whether err is a rate-limit signal — either a
// typed *APIError with HTTP 429 or the ErrRateLimited sentinel.
func IsRateLimited(err error) bool {
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	return api.HTTPStatus == 429
}

// IsConfigError reports whether err is non-recoverable operator
// misconfiguration (bad credential, unknown model, exhausted credit).
// Superset of IsAuthError + ErrInvalidConfig + ErrInsufficientCredits +
// HTTP 404.
func IsConfigError(err error) bool {
	if errors.Is(err, ErrInvalidConfig) || errors.Is(err, ErrInsufficientCredits) {
		return true
	}
	if IsAuthError(err) {
		return true
	}
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	return api.HTTPStatus == 404 || api.HTTPStatus == 402
}

// IsRetryable reports whether an HTTP status should be retried
// transparently by the Client retry loop. 408 is OpenRouter's request
// timeout and 502/503/504 mean the routed model or provider was
// momentarily unavailable — all safe to re-issue. 429 is handled
// separately (see Client.shouldRetry) because it needs Retry-After.
func IsRetryable(httpStatus int) bool {
	switch httpStatus {
	case 408, 502, 503, 504:
		return true
	}
	return false
}

// codeString renders wireError.Code, which the API documents as a number
// but OpenAI-compatible upstreams sometimes forward as a string.
func codeString(code any) string {
	switch v := code.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		if v == math.Trunc(v) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toAPIError converts an error envelope into the typed shape. httpStatus
// is the transport status; when the envelope carries its own numeric code
// and the transport said 200 (the streamed-failure case) the envelope code
// wins, because that is the real outcome.
func (w *wireError) toAPIError(httpStatus int) *APIError {
	if w == nil {
		return nil
	}
	code := codeString(w.Code)
	if httpStatus < 400 {
		if n, err := strconv.Atoi(code); err == nil && n >= 400 {
			httpStatus = n
		}
	}
	return &APIError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    w.Message,
	}
}

// classifyAPIError lifts a typed *APIError onto the package sentinels so
// the adapter's status mapping is driven by the taxonomy rather than by
// raw HTTP numbers. Unrecognised statuses are returned unchanged and end
// up as BOT_OP_FAILED.
//
// Wrapping uses two %w verbs on purpose: callers must be able to reach the
// sentinel with errors.Is AND the *APIError (for http_status / vendor_code)
// with errors.As. A %v on the second operand would flatten the type away.
func classifyAPIError(api *APIError) error {
	if api == nil {
		return nil
	}
	switch api.HTTPStatus {
	case 402:
		return fmt.Errorf("%w: %w", ErrInsufficientCredits, api)
	case 429:
		return fmt.Errorf("%w: %w", ErrRateLimited, api)
	case 400:
		// OpenRouter folds "prompt is longer than the context window" into
		// a generic 400, so the message text is the only discriminator.
		if isContextLengthMessage(api.Message) {
			return fmt.Errorf("%w: %w", ErrContextTooLarge, api)
		}
	}
	return api
}

// isContextLengthMessage matches the vendor phrasings that mean "the input
// did not fit". Same substring set as libs/eino.translateError.
func isContextLengthMessage(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "context length") ||
		strings.Contains(m, "context window") ||
		strings.Contains(m, "maximum context") ||
		strings.Contains(m, "too many tokens")
}

// parseRetryAfter reads the `Retry-After` header. OpenRouter sends it as a
// whole number of seconds; an HTTP-date form is also accepted by the spec
// and handled here for completeness. Returns 0 when absent or unparsable.
func parseRetryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs <= 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	if t, err := time.Parse(time.RFC1123, v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}
