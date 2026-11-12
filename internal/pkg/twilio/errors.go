package twilio

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrDecodeResponse signals that a 2xx Twilio response carried a body
// the client could not parse. The adapter factory maps it to
// SMS_SERIALIZE_FAILED.
var ErrDecodeResponse = errors.New("twilio: failed to decode response")

// APIError is the typed shape of a Twilio REST error response.
//
// Twilio returns errors as a JSON envelope (documented at
// https://www.twilio.com/docs/usage/twilios-response#error-response-format):
//
//	{"code": 21211, "message": "Invalid 'To' Phone Number",
//	 "more_info": "https://www.twilio.com/docs/errors/21211", "status": 400}
//
// HTTPStatus captures the actual HTTP status code separately because
// Twilio sometimes returns a 200 wrapping an error envelope (rare, but
// the contract is "trust the body when present, fall back to HTTP
// otherwise").
type APIError struct {
	HTTPStatus int    `json:"-"`
	Code       int    `json:"code"`
	Message    string `json:"message"`
	MoreInfo   string `json:"more_info"`
	Status     int    `json:"status"`
}

// Error formats the error without leaking the request payload or any
// credentials. Safe to interpolate into log lines.
func (e *APIError) Error() string {
	if e == nil {
		return "twilio: <nil APIError>"
	}
	if e.Code != 0 {
		return fmt.Sprintf("twilio: http %d code %d: %s", e.HTTPStatus, e.Code, e.Message)
	}
	return fmt.Sprintf("twilio: http %d: %s", e.HTTPStatus, e.Message)
}

const apiErrorMessageCap = 256

// parseAPIError decodes Twilio's standard error envelope from a non-2xx
// response. On JSON-decode failure (HTML 5xx pages, truncated bodies)
// it returns an APIError carrying the raw body truncated to a fixed
// cap so log lines never balloon.
func parseAPIError(httpStatus int, body []byte) *APIError {
	out := &APIError{HTTPStatus: httpStatus}
	if len(body) == 0 {
		out.Message = "<empty body>"
		return out
	}
	if err := json.Unmarshal(body, out); err == nil && (out.Code != 0 || out.Message != "") {
		out.HTTPStatus = httpStatus
		return out
	}
	// Decode failed or body wasn't a Twilio error envelope. Surface the
	// raw body for operator triage, truncated.
	msg := string(body)
	if len(msg) > apiErrorMessageCap {
		msg = msg[:apiErrorMessageCap]
	}
	out.Message = msg
	return out
}

// Twilio numeric error codes used by the classifiers below.
const (
	twilioAuthError                = 20003
	twilioAccountSuspended         = 20005
	twilioTooManyRequests          = 20429
	twilioAccountNotActive         = 20404
	twilioPhoneNotOwned            = 21606
	twilioMessagingServiceNotFound = 21659

	// twilioCountryNotEnabled — geo permissions for the destination
	// region are not enabled on the Twilio account. Operator must
	// enable in Twilio Console → Messaging → Geo Permissions.
	twilioCountryNotEnabled = 21408

	// twilioToUnreachable — the From number / Messaging Service does
	// not have a route to the destination. Operator must add an
	// internationally-capable sender or switch to a Messaging Service.
	twilioToUnreachable = 21612
)

// IsAuthError reports whether err looks like an authentication failure.
// True for HTTP 401/403, or for Twilio body codes covering auth /
// account-suspended / account-not-active.
func IsAuthError(err error) bool {
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	if api.HTTPStatus == 401 || api.HTTPStatus == 403 {
		return true
	}
	switch api.Code {
	case twilioAuthError, twilioAccountSuspended, twilioAccountNotActive, twilioTooManyRequests:
		return true
	}
	return false
}

// IsConfigError reports whether err looks like an operator
// misconfiguration: bad creds, bad From number, missing messaging
// service, OR the destination is not reachable from this account's
// configured senders (geo permissions / routing). Superset of
// IsAuthError. Callers treat this as non-recoverable: do not retry,
// surface to the operator.
func IsConfigError(err error) bool {
	if IsAuthError(err) {
		return true
	}
	var api *APIError
	if !errors.As(err, &api) {
		return false
	}
	switch api.Code {
	case twilioPhoneNotOwned, twilioMessagingServiceNotFound,
		twilioCountryNotEnabled, twilioToUnreachable:
		return true
	}
	return false
}

// IsRetryable reports whether an HTTP status code should be retried.
// True for 502/503/504. Twilio 429 (rate-limited) is NOT retried — we
// surface it so the caller can apply business-level backoff.
func IsRetryable(httpStatus int) bool {
	switch httpStatus {
	case 502, 503, 504:
		return true
	}
	return false
}
