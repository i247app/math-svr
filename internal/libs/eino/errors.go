package eino

import (
	"errors"
	"fmt"

	openaisdk "github.com/meguminnnnnnnnn/go-openai"
	"google.golang.org/genai"
)

// Sentinel errors surfaced by the Client. Each maps to a distinct
// MathError(BOT_*) code in the adapter layer; see adapter/bot/errors.go.
// The taxonomy is identical to libs/langchain so mapEinoError and
// mapLangChainError stay symmetrical.
var (
	// ErrDecodeResponse signals that an LLM call returned 2xx but the
	// body the client expected was missing or unparsable. Adapter maps
	// to BOT_SERIALIZE_FAILED.
	ErrDecodeResponse = errors.New("eino: failed to decode response")

	// ErrContextTooLarge signals that the input exceeded the backend's
	// context window. Detected from vendor-specific error strings.
	// Adapter maps to BOT_CONTEXT_TOO_LARGE.
	ErrContextTooLarge = errors.New("eino: context window exceeded")

	// ErrUnsupportedOp signals that the configured backend does not
	// support the requested capability. Adapter maps to BOT_UNSUPPORTED_OP.
	ErrUnsupportedOp = errors.New("eino: operation not supported by backend")

	// ErrRateLimited signals a vendor 429. Adapter maps to BOT_RATE_LIMITED.
	ErrRateLimited = errors.New("eino: rate limit exceeded")
)

// APIError is the typed shape of an upstream vendor error response. The
// client lifts the two primary vendor SDK error types (google genai,
// go-openai) into this shape in translateError; other vendors surface
// plain errors that are classified by substring only.
type APIError struct {
	Backend    Backend
	HTTPStatus int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e == nil {
		return "eino: <nil APIError>"
	}
	if e.Code != "" {
		return fmt.Sprintf("eino[%s]: http %d code %s: %s",
			e.Backend, e.HTTPStatus, e.Code, e.Message)
	}
	if e.HTTPStatus != 0 {
		return fmt.Sprintf("eino[%s]: http %d: %s",
			e.Backend, e.HTTPStatus, e.Message)
	}
	return fmt.Sprintf("eino[%s]: %s", e.Backend, e.Message)
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

// liftVendorError converts the vendor SDK error types the eino-ext
// components surface into a typed *APIError, preserving the original via
// %w so errors.Is/As keep working. Returns nil when err carries no
// recognisable vendor shape.
func liftVendorError(backend Backend, err error) error {
	var gErr *genai.APIError
	if errors.As(err, &gErr) {
		return fmt.Errorf("%w", &APIError{
			Backend:    backend,
			HTTPStatus: gErr.Code,
			Code:       gErr.Status,
			Message:    gErr.Message,
		})
	}

	var oErr *openaisdk.APIError
	if errors.As(err, &oErr) {
		code := ""
		if oErr.Type != "" {
			code = oErr.Type
		}
		return fmt.Errorf("%w", &APIError{
			Backend:    backend,
			HTTPStatus: oErr.HTTPStatusCode,
			Code:       code,
			Message:    oErr.Message,
		})
	}

	var oReqErr *openaisdk.RequestError
	if errors.As(err, &oReqErr) {
		return fmt.Errorf("%w", &APIError{
			Backend:    backend,
			HTTPStatus: oReqErr.HTTPStatusCode,
			Message:    oReqErr.Error(),
		})
	}

	return nil
}
