// Package firebase is a thin, domain-agnostic wrapper over
// firebase.google.com/go/v4. It owns nothing but the Firebase Admin SDK
// handle and the FCM messaging client; callers (the notification adapter)
// translate its results into domain shapes and MathErrors.
//
// Nothing here knows about ma_notifications, devices, or users — it only
// knows how to push a payload to a set of FCM registration tokens.
package firebase

import (
	"context"
	"errors"
	"fmt"
	"os"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// ErrInvalidConfig is returned by NewClient when required configuration is
// missing. The adapter factory maps it to NOTIFICATION_CONFIG_INVALID.
var ErrInvalidConfig = errors.New("firebase: invalid config")

// credTypeServiceAccount is the only credential type the Firebase Admin SDK
// accepts. We pass it explicitly to option.WithAuthCredentialsJSON so the
// library rejects any JSON that is not a service account — the validation the
// deprecated option.WithCredentialsJSON skipped (the documented security risk).
const credTypeServiceAccount = "service_account"

// Config holds the credentials needed to build a Firebase Admin client.
// CredentialsFile points at a service-account JSON (kept under keys/ — a
// forbidden read for agents). ProjectID is optional; when empty the SDK
// derives it from the credentials file.
type Config struct {
	CredentialsFile string
	ProjectID       string
}

// Client wraps the FCM messaging client. Safe for concurrent use after
// construction.
type Client struct {
	messaging *messaging.Client
}

// PushPayload is the provider-agnostic push body. Data carries the optional
// custom key/value map delivered alongside the visible notification (used by
// the mobile client to deep-link via action_type / action_data).
type PushPayload struct {
	Title string
	Body  string
	Data  map[string]string
}

// MulticastResult summarises a fan-out send. InvalidTokens lists the
// registration tokens FCM rejected as unregistered / malformed, so the
// caller can prune them from ma_devices.
type MulticastResult struct {
	SuccessCount  int
	FailureCount  int
	InvalidTokens []string
}

// NewClient builds a messaging client from cfg.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.CredentialsFile == "" {
		return nil, fmt.Errorf("%w: credentials file is required", ErrInvalidConfig)
	}

	creds, err := os.ReadFile(cfg.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("firebase: read credentials: %w", err)
	}

	app, err := fb.NewApp(ctx, &fb.Config{ProjectID: cfg.ProjectID},
		option.WithAuthCredentialsJSON(option.CredentialsType(credTypeServiceAccount), creds))
	if err != nil {
		return nil, fmt.Errorf("firebase: init app: %w", err)
	}

	msg, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase: init messaging: %w", err)
	}

	return &Client{messaging: msg}, nil
}

// SendMulticast pushes p to every token in tokens via a single FCM
// multicast call. It returns per-token success/failure counts plus the
// subset of tokens FCM reported as dead (unregistered or invalid), which the
// caller should delete. An empty tokens slice is a no-op (zeroed result).
func (c *Client) SendMulticast(ctx context.Context, tokens []string, p PushPayload) (*MulticastResult, error) {
	if len(tokens) == 0 {
		return &MulticastResult{}, nil
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: p.Title,
			Body:  p.Body,
		},
		Data: p.Data,
	}

	resp, err := c.messaging.SendEachForMulticast(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("firebase: send multicast: %w", err)
	}

	result := &MulticastResult{
		SuccessCount: resp.SuccessCount,
		FailureCount: resp.FailureCount,
	}
	for i, r := range resp.Responses {
		if r.Success || r.Error == nil {
			continue
		}
		// Tokens FCM no longer recognises should be pruned by the caller.
		if messaging.IsUnregistered(r.Error) || messaging.IsInvalidArgument(r.Error) {
			result.InvalidTokens = append(result.InvalidTokens, tokens[i])
		}
	}
	return result, nil
}
