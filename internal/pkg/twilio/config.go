// Package twilio is the bin-svr libs wrapper around Twilio's REST API.
// It owns the http_client.Client lifecycle (auth, timeouts, retry) and
// exposes a narrow resource-typed surface (Messages.Create, ...) that
// the adapter layer (`internal/adapter/sms`) consumes.
//
// This package never imports from `internal/domain` or
// `internal/adapter` — it is the lowest layer, mirroring the role of
// `internal/libs/redis` for the cache adapter and `internal/libs/email`
// for the email adapter. Errors returned here are plain wrapped errors
// or *APIError values; the adapter factory translates them into
// binbaseError(SMS_CONNECT_FAILED / SMS_CONFIG_INVALID / SMS_OP_FAILED /
// SMS_SERIALIZE_FAILED).
//
// Secret handling: cfg.AuthToken is read once at construction and held
// in Client.authToken solely so http_client.WithBasicAuth can encode it
// into the outbound Authorization header. It is NEVER logged, never
// included in error strings, and never embedded in panics. Grep for
// `AuthToken` in this package and audits should find only:
//   - the struct field definition,
//   - the WithBasicAuth call site in client.go,
//   - test fixtures.
package twilio

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ErrInvalidConfig is the sentinel wrapped by Config.Validate. The
// adapter factory uses errors.Is to distinguish "config is wrong"
// (→ SMS_CONFIG_INVALID) from "config is fine but the network or
// upstream said no" (→ SMS_CONNECT_FAILED).
var ErrInvalidConfig = errors.New("twilio: invalid config")

// Config holds the parameters for the Twilio REST client.
//
// AccountSID + AuthToken are the production credential pair. AuthToken
// is sensitive — see package doc.
//
// From and MessagingServiceSID are the default senders. At least one of
// (From, MessagingServiceSID) must resolve at send time, either from
// this config or from the per-call Message.From. Validate() does NOT
// require either at boot, because a caller that always supplies
// Message.From is a legitimate setup.
type Config struct {
	// AccountSID identifies the Twilio account. Matches ^AC[a-f0-9]{32}$.
	AccountSID string

	// AuthToken is the long-lived secret for HTTP Basic auth. NEVER log.
	AuthToken string

	// BaseURL overrides the Twilio REST endpoint. Empty resolves to
	// "https://api.twilio.com". Non-empty must start with "https://".
	BaseURL string

	// From is the E.164 sender phone provisioned in the Twilio account.
	// Optional; falls back to MessagingServiceSID, then per-message
	// Message.From at send time.
	From string

	// MessagingServiceSID identifies a Messaging Service (geo-aware
	// routing pool). Twilio recommends this for production.
	MessagingServiceSID string

	// Timeout bounds a single HTTP request to Twilio. applyDefaults sets
	// 10s when zero.
	Timeout time.Duration

	// MaxRetries is the per-request retry budget for 5xx responses.
	// applyDefaults sets 2 when zero. Twilio 429 is NOT retried — see
	// errors.go.
	MaxRetries int

	// RetryDelay is the linear sleep between retry attempts.
	// applyDefaults sets 250ms when zero.
	RetryDelay time.Duration

	// RequireAtBoot governs NewClient's startup probe. true → fail fast
	// on a bad credential. false → log a warning and continue so dev
	// environments without Twilio still boot.
	RequireAtBoot bool
}

// DefaultBaseURL is the public Twilio REST endpoint.
const DefaultBaseURL = "https://api.twilio.com"

var accountSIDPattern = regexp.MustCompile(`^AC[a-f0-9]{32}$`)

// Validate enforces the credential format and the https-only invariant.
// Returns errors wrapped with ErrInvalidConfig; the adapter factory
// translates that sentinel into binbaseError(SMS_CONFIG_INVALID).
func (c Config) Validate() error {
	if c.AccountSID == "" {
		return fmt.Errorf("%w: AccountSID is required", ErrInvalidConfig)
	}
	if !accountSIDPattern.MatchString(c.AccountSID) {
		return fmt.Errorf("%w: AccountSID must match ^AC[a-f0-9]{32}$", ErrInvalidConfig)
	}
	if c.AuthToken == "" {
		return fmt.Errorf("%w: AuthToken is required", ErrInvalidConfig)
	}
	if len(c.AuthToken) < 32 {
		return fmt.Errorf("%w: AuthToken length must be >= 32", ErrInvalidConfig)
	}
	if c.BaseURL != "" && !strings.HasPrefix(c.BaseURL, "https://") {
		return fmt.Errorf("%w: BaseURL must use https:// (got %q)", ErrInvalidConfig, c.BaseURL)
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
// AFTER Validate (which only enforces hard invariants) so callers can
// leave the transport-tuning fields blank.
func (c *Config) applyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.Timeout == 0 {
		c.Timeout = 10 * time.Second
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 2
	}
	if c.RetryDelay == 0 {
		c.RetryDelay = 250 * time.Millisecond
	}
}

// MaskPhone returns a log-safe rendering of an E.164 phone number that
// preserves the leading "+", the country-code prefix, and the last four
// digits while replacing the middle with "***". Inputs too short to
// mask are returned unchanged — empty input returns empty.
//
// Example: "+15551234567" → "+1***4567".
//
// Lives in the libs package because phone-PII masking is a Twilio-payload
// concern; the adapter layer re-exports it via package-level usage.
func MaskPhone(p string) string {
	if p == "" {
		return ""
	}
	// Need at least "+" + 1 cc digit + 4 trailing digits + 1 hidden char.
	if len(p) < 7 {
		return p
	}
	if !strings.HasPrefix(p, "+") {
		return p
	}
	// Keep the "+", the first cc digit (1–3 chars commonly, but we keep 1
	// for safety since the country code length isn't trivially derivable
	// from the prefix alone), and the last 4 digits.
	keepHead := 2 // "+" + first digit
	keepTail := 4
	if len(p) <= keepHead+keepTail {
		return p
	}
	return p[:keepHead] + "***" + p[len(p)-keepTail:]
}
