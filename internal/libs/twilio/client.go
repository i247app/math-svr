package twilio

import (
	"context"
	"fmt"
	"net/url"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/http_client"
)

// Client is the Twilio REST client. It owns the configured
// *http_client.Client and exposes resource services (Messages, ...)
// that hide the wire shape from callers.
//
// The client is safe for concurrent use; http_client.Client shares a
// single net/http.Client whose Transport pools connections internally.
//
// Lifecycle: there is nothing to Close — net/http.Transport manages its
// own connection pool. The struct is constructed once at boot and
// retained on the SMS adapter for the process lifetime.
type Client struct {
	// accountSID and authToken are captured here so resource services
	// (messages.go) can construct REST paths and the HTTP client can
	// encode the Authorization header. authToken is NEVER logged.
	accountSID string
	authToken  string

	// defaultFrom / messagingServiceSID record the boot-time defaults.
	// Resource services may also consult them when a per-call value is
	// absent; the TwilioProvider in adapter/sms is the canonical caller
	// of that fallback.
	defaultFrom         string
	messagingServiceSID string

	// http is the wired shared HTTP client. Built with NO logging
	// interceptor — see NewClient.
	http *http_client.Client

	// Messages is the Messages resource service.
	Messages *MessagesService
}

// AccountSID returns the Twilio account identifier the client was
// constructed with. Safe to log.
func (c *Client) AccountSID() string { return c.accountSID }

// DefaultFrom returns the configured default From phone. Safe to log
// when masked.
func (c *Client) DefaultFrom() string { return c.defaultFrom }

// MessagingServiceSID returns the configured Messaging Service SID.
// Safe to log.
func (c *Client) MessagingServiceSID() string { return c.messagingServiceSID }

// NewClient validates cfg, builds the underlying http_client.Client with
// the right auth + retry policy, and (when cfg.RequireAtBoot is true)
// runs a boot-time probe against the Twilio account-fetch endpoint to
// fail fast on bad credentials.
//
// The inner http_client.Client is constructed with an EXPLICIT
// replacement interceptor list containing only the timing interceptor.
// The default logging interceptor would capture the outbound
// Authorization header and the form-encoded request body (which
// contains To= in plain text); both are sensitive. Adapter-level
// structured logging in `internal/adapter/sms` covers what we need.
//
// Errors:
//   - cfg.Validate failure → wraps ErrInvalidConfig (adapter factory
//     maps to binbaseError(SMS_CONFIG_INVALID)).
//   - probe failure under RequireAtBoot=true → wraps the underlying
//     error (adapter factory maps to binbaseError(SMS_CONNECT_FAILED)).
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.applyDefaults()

	httpClient := http_client.NewClient(
		http_client.WithBaseURL(cfg.BaseURL),
		http_client.WithBasicAuth(cfg.AccountSID, cfg.AuthToken),
		http_client.WithTimeout(cfg.Timeout),
		http_client.WithUserAgent("binbase-svr/twilio (REST)"),
		http_client.WithAccept("application/json"),
		http_client.WithInterceptors(http_client.NewTimingInterceptor()),
		http_client.WithRetry(cfg.MaxRetries, cfg.RetryDelay, 502, 503, 504),
	)

	c := &Client{
		accountSID:          cfg.AccountSID,
		authToken:           cfg.AuthToken,
		defaultFrom:         cfg.From,
		messagingServiceSID: cfg.MessagingServiceSID,
		http:                httpClient,
	}
	c.Messages = &MessagesService{c: c}

	if cfg.RequireAtBoot {
		if err := c.probe(ctx); err != nil {
			return nil, fmt.Errorf("twilio: boot probe failed: %w", err)
		}
	} else {
		log := logger.From(ctx)
		log.Infof("twilio: client ready (account=%s, base=%s, probe-skipped)", cfg.AccountSID, cfg.BaseURL)
	}

	return c, nil
}

// probe makes a single GET against the account-fetch endpoint to verify
// credentials and connectivity. Used only when RequireAtBoot=true.
//
// Twilio's REST API has no dedicated health endpoint; fetching the
// account resource is the documented credential check (a 401/403 means
// the AuthToken is wrong, a 200 means the credentials work).
func (c *Client) probe(ctx context.Context) error {
	path := fmt.Sprintf("/2010-04-01/Accounts/%s.json", url.PathEscape(c.accountSID))
	resp, err := c.http.Get(ctx, path)
	if err != nil {
		return fmt.Errorf("transport: %w", err)
	}
	if !resp.IsSuccess() {
		return parseAPIError(resp.StatusCode, resp.Bytes())
	}
	return nil
}
