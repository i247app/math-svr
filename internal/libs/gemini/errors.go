package gemini

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
// The taxonomy matches the sibling libs packages so mapGeminiError stays
// symmetrical with its four siblings — plus ErrContentBlocked, which has
// no analogue on the OpenAI-shaped providers.
var (
	// ErrDecodeResponse signals that a call returned 2xx but the body the
	// client expected was missing or unparsable. → BOT_SERIALIZE_FAILED.
	ErrDecodeResponse = errors.New("gemini: failed to decode response")

	// ErrContextTooLarge signals that the input exceeded the model's
	// context window. → BOT_CONTEXT_TOO_LARGE.
	ErrContextTooLarge = errors.New("gemini: context window exceeded")

	// ErrUnsupportedOp signals that the requested capability is not
	// configured. → BOT_UNSUPPORTED_OP.
	ErrUnsupportedOp = errors.New("gemini: operation not supported")

	// ErrRateLimited signals a throttling 429 (RESOURCE_EXHAUSTED from a
	// per-minute limit) — transient. → BOT_RATE_LIMITED.
	ErrRateLimited = errors.New("gemini: rate limit exceeded")

	// ErrQuotaExhausted signals a 429 that is a DAILY quota or a billing
	// stop rather than a burst limit. It shares the status code with
	// throttling but will not recover within the call.
	// → BOT_CONFIG_INVALID.
	ErrQuotaExhausted = errors.New("gemini: quota exhausted")

	// ErrContentBlocked signals that a safety / recitation filter refused
	// the request. Unique to this provider among the four; retrying is
	// pointless because the same prompt produces the same block.
	// → BOT_CONTENT_BLOCKED.
	ErrContentBlocked = errors.New("gemini: content blocked by safety filter")
)

// APIError is the typed shape of a Gemini error response.
//
// Message carries the vendor's own error text — never the request prompt
// and never the raw response body, so it is safe to log and to surface in
// the MathError debug field.
type APIError struct {
	HTTPStatus int
	// Code is the vendor code rendered as a string. Gemini returns either
	// a number equal to the HTTP status (google.rpc style) or a snake_case
	// identifier (newer style); both land here.
	Code string
	// Status is the google.rpc status string when present, e.g.
	// "RESOURCE_EXHAUSTED", "INVALID_ARGUMENT", "UNAUTHENTICATED".
	Status  string
	Message string
	// RetryAfter is the RetryInfo hint from error.details, or the
	// Retry-After header. Zero when absent.
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e == nil {
		return "gemini: <nil APIError>"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "gemini: http %d", e.HTTPStatus)
	if e.Status != "" {
		fmt.Fprintf(&b, " status %s", e.Status)
	}
	if e.Code != "" && e.Code != strconv.Itoa(e.HTTPStatus) {
		fmt.Fprintf(&b, " code %s", e.Code)
	}
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	}
	return b.String()
}

// IsAuthError reports whether err is an authentication / authorization
// failure. 401 is a missing or invalid key; 403 is a key without
// permission for the resource.
func IsAuthError(err error) bool {
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	if api.HTTPStatus == 401 || api.HTTPStatus == 403 {
		return true
	}
	switch api.Status {
	case "UNAUTHENTICATED", "PERMISSION_DENIED":
		return true
	}
	return false
}

// IsRateLimited reports whether err is transient throttling. Deliberately
// FALSE for a daily-quota 429 — that is ErrQuotaExhausted.
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
	return api.HTTPStatus == 429 && !isQuotaExhaustion(api)
}

// IsConfigError reports whether err is non-recoverable operator
// misconfiguration (bad key, no permission, unknown model, disabled
// billing, exhausted daily quota).
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
	if api.HTTPStatus == 404 {
		return true
	}
	// FAILED_PRECONDITION is Gemini's "billing is not enabled" / "not
	// available in your region" signal — an operator problem, not a
	// transient one.
	return api.Status == "FAILED_PRECONDITION"
}

// IsContentBlocked reports whether err is a safety / recitation refusal.
func IsContentBlocked(err error) bool { return errors.Is(err, ErrContentBlocked) }

// IsRetryable reports whether an HTTP status should be retried
// transparently. 408/500/502/503/504 are transport or upstream faults. 429
// is handled separately by retryDelayFor because burst throttling and
// daily quota share that status.
func IsRetryable(httpStatus int) bool {
	switch httpStatus {
	case 408, 500, 502, 503, 504:
		return true
	}
	return false
}

// isQuotaExhaustion distinguishes a daily-quota / billing 429 from burst
// throttling. Gemini does not give these separate codes, so the message
// text is the only discriminator the API offers.
func isQuotaExhaustion(api *APIError) bool {
	if api == nil {
		return false
	}
	m := strings.ToLower(api.Message)
	switch {
	case strings.Contains(m, "quota exceeded"),
		strings.Contains(m, "daily limit"),
		strings.Contains(m, "billing"),
		strings.Contains(m, "free tier"):
		return true
	}
	return strings.EqualFold(api.Code, "quota_exceeded")
}

func isContextLengthMessage(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "context length") ||
		strings.Contains(m, "context window") ||
		strings.Contains(m, "maximum context") ||
		strings.Contains(m, "token count") && strings.Contains(m, "exceeds") ||
		strings.Contains(m, "too many tokens") ||
		strings.Contains(m, "input token count")
}

// codeString renders wireError.Code, which is a number in the google.rpc
// shape and a snake_case string in the newer one.
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

// retryDelay pulls the RetryInfo hint out of error.details. Google encodes
// it as a Go-style duration string ("17s").
func (w *wireError) retryDelay() time.Duration {
	if w == nil {
		return 0
	}
	for _, d := range w.Details {
		if !strings.Contains(d.Type, "RetryInfo") || d.RetryDelay == "" {
			continue
		}
		if parsed, err := time.ParseDuration(d.RetryDelay); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 0
}

// toAPIError converts an error envelope into the typed shape. When the
// transport status is unknown (a 200 carrying an error body) the envelope's
// numeric code stands in.
func (w *wireError) toAPIError(httpStatus int, retryAfter time.Duration) *APIError {
	if w == nil {
		return nil
	}
	code := codeString(w.Code)
	if httpStatus < 400 {
		if n, err := strconv.Atoi(code); err == nil && n >= 400 {
			httpStatus = n
		}
	}
	if hint := w.retryDelay(); hint > 0 {
		retryAfter = hint
	}
	return &APIError{
		HTTPStatus: httpStatus,
		Code:       code,
		Status:     w.Status,
		Message:    w.Message,
		RetryAfter: retryAfter,
	}
}

// classifyAPIError lifts a typed *APIError onto the package sentinels so
// the adapter's status mapping is driven by the taxonomy rather than by
// raw HTTP numbers. Unrecognised statuses are returned unchanged and end
// up as BOT_OP_FAILED.
//
// Wrapping uses two %w verbs on purpose: callers must be able to reach the
// sentinel with errors.Is AND the *APIError (for http_status / status)
// with errors.As. A %v on the second operand would flatten the type away.
func classifyAPIError(api *APIError) error {
	if api == nil {
		return nil
	}

	if api.HTTPStatus == 429 || api.Status == "RESOURCE_EXHAUSTED" {
		// Daily quota before burst throttling: retrying the former burns
		// the whole budget on a call that cannot succeed today.
		if isQuotaExhaustion(api) {
			return fmt.Errorf("%w: %w", ErrQuotaExhausted, api)
		}
		return fmt.Errorf("%w: %w", ErrRateLimited, api)
	}

	// Gemini reports an oversized prompt as a plain INVALID_ARGUMENT, so
	// the message text is the only discriminator.
	if api.HTTPStatus == 400 && isContextLengthMessage(api.Message) {
		return fmt.Errorf("%w: %w", ErrContextTooLarge, api)
	}

	return api
}

// parseRetryAfter reads a Retry-After header. Gemini normally puts its
// hint in error.details instead, but a fronting proxy may set the header.
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
