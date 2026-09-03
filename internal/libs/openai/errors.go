package openai

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Sentinel errors surfaced by the Client. Each maps to a distinct
// MathError(BOT_*) code in the adapter layer; see adapter/bot/errors.go.
// The taxonomy matches the sibling libs packages so mapOpenAIError stays
// symmetrical with its three siblings.
var (
	// ErrDecodeResponse signals that a call returned 2xx but the body the
	// client expected was missing or unparsable. → BOT_SERIALIZE_FAILED.
	ErrDecodeResponse = errors.New("openai: failed to decode response")

	// ErrContextTooLarge signals that the input exceeded the model's
	// context window. → BOT_CONTEXT_TOO_LARGE.
	ErrContextTooLarge = errors.New("openai: context window exceeded")

	// ErrUnsupportedOp signals that the requested capability is not
	// configured. → BOT_UNSUPPORTED_OP.
	ErrUnsupportedOp = errors.New("openai: operation not supported")

	// ErrRateLimited signals a throttling 429 — transient, worth retrying.
	// → BOT_RATE_LIMITED.
	ErrRateLimited = errors.New("openai: rate limit exceeded")

	// ErrQuotaExhausted signals a BILLING 429, which OpenAI reports as
	// type=rate_limit_error + code=credit_balance_exhausted (historically
	// insufficient_quota). It shares the status code with plain throttling
	// but is the opposite situation: it will never succeed on retry and
	// needs an operator to add credit. → BOT_CONFIG_INVALID.
	ErrQuotaExhausted = errors.New("openai: quota / credit balance exhausted")
)

// APIError is the typed shape of an OpenAI error response.
//
// Message carries the vendor's own error text — never the request prompt
// and never the raw response body, so it is safe to log and to surface in
// the MathError debug field.
type APIError struct {
	HTTPStatus int
	// Type is the vendor category, e.g. "invalid_request_error",
	// "rate_limit_error", "invalid_authentication_error".
	Type string
	// Code is the vendor's string identifier, e.g.
	// "credit_balance_exhausted", "context_length_exceeded". Note this is
	// a STRING on OpenAI, unlike OpenRouter's numeric code.
	Code string
	// Param names the offending request field on a 400, when reported.
	Param   string
	Message string
	// RetryAfter is the parsed `Retry-After` header. Zero when absent.
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e == nil {
		return "openai: <nil APIError>"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "openai: http %d", e.HTTPStatus)
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

// IsAuthError reports whether err is an authentication / authorization
// failure. 401 is a bad or missing key; 403 on OpenAI means the request
// came from an unsupported country or region, which is equally an
// operator problem rather than a transient one.
func IsAuthError(err error) bool {
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	return api.HTTPStatus == 401 || api.HTTPStatus == 403
}

// IsRateLimited reports whether err is throttling. Deliberately FALSE for
// a billing 429 — that is ErrQuotaExhausted, which must not be retried.
func IsRateLimited(err error) bool {
	if errors.Is(err, ErrQuotaExhausted) {
		return false
	}
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	return api.HTTPStatus == 429 && !isQuotaCode(api.Code)
}

// IsConfigError reports whether err is non-recoverable operator
// misconfiguration (bad credential, geo block, unknown model, exhausted
// credit). Superset of IsAuthError + ErrInvalidConfig + ErrQuotaExhausted
// + HTTP 404.
func IsConfigError(err error) bool {
	if errors.Is(err, ErrInvalidConfig) || errors.Is(err, ErrQuotaExhausted) {
		return true
	}
	if IsAuthError(err) {
		return true
	}
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	return api.HTTPStatus == 404 || isQuotaCode(api.Code)
}

// IsRetryable reports whether an HTTP status should be retried
// transparently. 408 is a request timeout; 500/502/503/504 are upstream
// faults or an overloaded model. 429 is handled separately by
// Client.shouldRetry because throttling and billing share that status.
func IsRetryable(httpStatus int) bool {
	switch httpStatus {
	case 408, 500, 502, 503, 504:
		return true
	}
	return false
}

// isQuotaCode recognises the vendor codes that mean "the account cannot
// pay for this call". OpenAI renamed insufficient_quota to
// credit_balance_exhausted; both are still observed in the wild.
func isQuotaCode(code string) bool {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "credit_balance_exhausted", "insufficient_quota", "billing_hard_limit_reached":
		return true
	}
	return false
}

// isContextLengthCode / isContextLengthMessage recognise "the input did
// not fit". OpenAI usually sets code=context_length_exceeded, but proxies
// and older deployments only produce prose, so both are checked.
func isContextLengthCode(code string) bool {
	return strings.EqualFold(strings.TrimSpace(code), "context_length_exceeded")
}

func isContextLengthMessage(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "context length") ||
		strings.Contains(m, "context window") ||
		strings.Contains(m, "maximum context") ||
		strings.Contains(m, "too many tokens")
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

	// Billing before throttling: both arrive as 429 and only the code
	// tells them apart, but retrying a billing failure burns the whole
	// budget for a call that can never succeed.
	if isQuotaCode(api.Code) {
		return fmt.Errorf("%w: %w", ErrQuotaExhausted, api)
	}
	if isContextLengthCode(api.Code) {
		return fmt.Errorf("%w: %w", ErrContextTooLarge, api)
	}

	switch api.HTTPStatus {
	case 429:
		return fmt.Errorf("%w: %w", ErrRateLimited, api)
	case 400:
		if isContextLengthMessage(api.Message) {
			return fmt.Errorf("%w: %w", ErrContextTooLarge, api)
		}
	}
	return api
}

// parseRetryAfter reads the `Retry-After` header. OpenAI sends whole
// seconds; the HTTP-date form is accepted for completeness. Returns 0 when
// absent or unparsable.
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
