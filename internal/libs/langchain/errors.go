package langchain

import (
	"errors"
	"fmt"
)

// Sentinel errors surfaced by the Client. Each maps to a distinct
// MathError(BOT_*) code in the adapter layer; see adapter/bot/errors.go.
var (
	// ErrDecodeResponse signals that an LLM call returned 2xx but the
	// body the client expected was missing or unparsable. Adapter maps
	// to BOT_SERIALIZE_FAILED.
	ErrDecodeResponse = errors.New("langchain: failed to decode response")

	// ErrContextTooLarge signals that the input exceeded the backend's
	// context window. Detected from vendor-specific error strings.
	// Adapter maps to BOT_CONTEXT_TOO_LARGE.
	ErrContextTooLarge = errors.New("langchain: context window exceeded")

	// ErrUnsupportedOp signals that the configured backend does not
	// support the requested capability (e.g. embeddings on a chat-only
	// vendor). Adapter maps to BOT_UNSUPPORTED_OP.
	ErrUnsupportedOp = errors.New("langchain: operation not supported by backend")

	// ErrRateLimited signals a vendor 429. Adapter maps to BOT_RATE_LIMITED.
	ErrRateLimited = errors.New("langchain: rate limit exceeded")
)

// APIError is the typed shape of an upstream vendor error response. Not
// every langchaingo backend produces one — vendor SDKs often surface
// plain errors. When the client cannot type-assert it builds a synthetic
// APIError with HTTPStatus=0 and Backend set so log lines stay uniform.
type APIError struct {
	Backend    Backend
	HTTPStatus int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e == nil {
		return "langchain: <nil APIError>"
	}
	if e.Code != "" {
		return fmt.Sprintf("langchain[%s]: http %d code %s: %s",
			e.Backend, e.HTTPStatus, e.Code, e.Message)
	}
	if e.HTTPStatus != 0 {
		return fmt.Sprintf("langchain[%s]: http %d: %s",
			e.Backend, e.HTTPStatus, e.Message)
	}
	return fmt.Sprintf("langchain[%s]: %s", e.Backend, e.Message)
}

// IsAuthError reports whether err is an authentication failure (401/403).
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
// misconfiguration (bad credential, unknown model, unsupported backend).
// Superset of IsAuthError + ErrInvalidConfig + HTTP 404.
func IsConfigError(err error) bool {
	if errors.Is(err, ErrInvalidConfig) {
		return true
	}
	if IsAuthError(err) {
		return true
	}
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	return api.HTTPStatus == 404
}

// IsRetryable reports whether an HTTP status code should be retried
// transparently by the Client retry loop. True for 502/503/504; 429 is
// surfaced (not auto-retried) so business-level backoff can apply.
func IsRetryable(httpStatus int) bool {
	switch httpStatus {
	case 502, 503, 504:
		return true
	}
	return false
}
