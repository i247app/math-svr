package deepseek

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Sentinel errors surfaced by the Client. Each maps to a distinct
// MathError(BOT_*) code in the adapter layer; see adapter/bot/errors.go.
// The taxonomy matches the sibling libs packages so mapDeepSeekError stays
// symmetrical with its four siblings.
var (
	// ErrDecodeResponse signals that a call returned 2xx but the body the
	// client expected was missing or unparsable. → BOT_SERIALIZE_FAILED.
	ErrDecodeResponse = errors.New("deepseek: failed to decode response")

	// ErrContextTooLarge signals that the input exceeded the model's
	// context window. → BOT_CONTEXT_TOO_LARGE.
	ErrContextTooLarge = errors.New("deepseek: context window exceeded")

	// ErrUnsupportedOp signals that the requested capability does not
	// exist on this provider (embeddings). → BOT_UNSUPPORTED_OP.
	ErrUnsupportedOp = errors.New("deepseek: operation not supported")

	// ErrRateLimited signals HTTP 429 — transient. → BOT_RATE_LIMITED.
	ErrRateLimited = errors.New("deepseek: rate limit exceeded")

	// ErrInsufficientBalance signals HTTP 402: the account is out of
	// credit. Non-recoverable within the call and needs an operator to top
	// up, so it rides with the config failures rather than a transient
	// code — the same posture as openrouter's 402 and openai's billing
	// 429. → BOT_CONFIG_INVALID.
	ErrInsufficientBalance = errors.New("deepseek: insufficient balance")

	// ErrContentFiltered signals the platform's content filter refused the
	// generation (finish_reason=content_filter). Retrying the same prompt
	// yields the same refusal. → BOT_CONTENT_BLOCKED.
	ErrContentFiltered = errors.New("deepseek: content filtered")

	// ErrServerOverloaded signals DeepSeek ran out of capacity. It arrives
	// either as HTTP 503 or, mid-generation, as
	// finish_reason=insufficient_system_resource. Transient.
	// → BOT_OP_FAILED.
	ErrServerOverloaded = errors.New("deepseek: server overloaded")
)

// APIError is the typed shape of a DeepSeek error response.
//
// Message carries the vendor's own error text — never the request prompt
// and never the raw response body, so it is safe to log and to surface in
// the MathError debug field.
type APIError struct {
	HTTPStatus int
	Type       string
	Code       string
	Param      string
	Message    string
	// RetryAfter is the parsed `Retry-After` header. Zero when absent.
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e == nil {
		return "deepseek: <nil APIError>"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "deepseek: http %d", e.HTTPStatus)
	if e.Type != "" {
		fmt.Fprintf(&b, " type %s", e.Type)
	}
	if e.Code != "" {
		fmt.Fprintf(&b, " code %s", e.Code)
	}
	if e.Param != "" {
		fmt.Fprintf(&b, " param %s", e.Param)
	}
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	}
	return b.String()
}

// IsAuthError reports whether err is an authentication failure. DeepSeek
// uses 401 for a bad key; it has no separate 403 path.
func IsAuthError(err error) bool {
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	return api.HTTPStatus == 401 || api.HTTPStatus == 403
}

// IsRateLimited reports whether err is throttling.
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
// misconfiguration (bad credential, unknown model, empty balance).
func IsConfigError(err error) bool {
	if errors.Is(err, ErrInvalidConfig) || errors.Is(err, ErrInsufficientBalance) {
		return true
	}
	if IsAuthError(err) {
		return true
	}
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	return api.HTTPStatus == 402 || api.HTTPStatus == 404
}

// IsRetryable reports whether an HTTP status should be retried
// transparently. Per DeepSeek's own guidance only 500 and 503 are worth a
// retry; 400/401/402/422/429 all need corrective action. 408 and 504 are
// included as transport-level timeouts a fronting proxy may produce.
func IsRetryable(httpStatus int) bool {
	switch httpStatus {
	case 408, 500, 502, 503, 504:
		return true
	}
	return false
}

// isContextLengthMessage recognises the phrasings that mean "the input did
// not fit". DeepSeek reports this as a plain 400, so the message text is
// the only discriminator.
func isContextLengthMessage(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "context length") ||
		strings.Contains(m, "context window") ||
		strings.Contains(m, "maximum context") ||
		strings.Contains(m, "too many tokens") ||
		strings.Contains(m, "exceeds the maximum")
}

// classifyAPIError lifts a typed *APIError onto the package sentinels so
// the adapter's status mapping is driven by the taxonomy rather than by
// raw HTTP numbers. Unrecognised statuses are returned unchanged and end
// up as BOT_OP_FAILED.
//
// Wrapping uses two %w verbs on purpose: callers must be able to reach the
// sentinel with errors.Is AND the *APIError (for http_status / vendor
// code) with errors.As. A %v on the second operand would flatten the type
// away.
func classifyAPIError(api *APIError) error {
	if api == nil {
		return nil
	}

	switch api.HTTPStatus {
	case 402:
		return fmt.Errorf("%w: %w", ErrInsufficientBalance, api)
	case 429:
		return fmt.Errorf("%w: %w", ErrRateLimited, api)
	case 503:
		return fmt.Errorf("%w: %w", ErrServerOverloaded, api)
	case 400, 422:
		// 422 is DeepSeek's "invalid parameters"; an oversized prompt can
		// land on either status, so both check the text.
		if isContextLengthMessage(api.Message) {
			return fmt.Errorf("%w: %w", ErrContextTooLarge, api)
		}
	}
	return api
}

// parseRetryAfter reads the `Retry-After` header. Returns 0 when absent or
// unparsable.
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
