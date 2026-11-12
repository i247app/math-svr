package email

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GoogleEmailConfig configures a Gmail API client that sends on behalf of a
// workspace user via domain-wide delegation. The service account JSON at
// CredentialsFile must be authorized to impersonate SenderEmail in the admin
// console with the gmail.send scope.
type GoogleEmailConfig struct {
	CredentialsFile string
	SenderEmail     string
}

// GoogleEmailClient wraps a gmail.Service scoped to impersonate a single sender.
type GoogleEmailClient struct {
	service *gmail.Service
	sender  string
}

func NewGoogleEmailClient(ctx context.Context, cfg GoogleEmailConfig) (*GoogleEmailClient, error) {
	if cfg.CredentialsFile == "" {
		return nil, fmt.Errorf("google email: credentials file is required")
	}
	if cfg.SenderEmail == "" {
		return nil, fmt.Errorf("google email: sender email is required")
	}

	data, err := os.ReadFile(cfg.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("google email: read credentials: %w", err)
	}

	jwtCfg, err := google.JWTConfigFromJSON(data, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("google email: parse credentials: %w", err)
	}
	jwtCfg.Subject = cfg.SenderEmail

	httpClient := jwtCfg.Client(ctx)
	svc, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("google email: create service: %w", err)
	}

	return &GoogleEmailClient{service: svc, sender: cfg.SenderEmail}, nil
}

// Sender returns the impersonated sender email used by this client.
func (c *GoogleEmailClient) Sender() string { return c.sender }

// SendRaw ships a RFC 2822 MIME payload through the Gmail API. Callers are
// responsible for supplying correct headers; the Gmail API resolves the actual
// "From" header based on the impersonated mailbox.
func (c *GoogleEmailClient) SendRaw(ctx context.Context, raw []byte) error {
	msg := &gmail.Message{Raw: encodeRawMessage(raw)}
	if _, err := c.service.Users.Messages.Send("me", msg).Context(ctx).Do(); err != nil {
		return fmt.Errorf("google email: send: %w", err)
	}
	return nil
}
