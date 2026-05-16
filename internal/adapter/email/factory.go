package email

import (
	"context"
	"fmt"

	lib_email "math-ai.com/math-ai/internal/libs/email"

	"math-ai.com/math-ai/internal/infrastructure/config"
)

// NewFromConfig builds an Adapter with a single provider based on cfg. Only the
// credentials for the selected provider need to be populated.
func NewFromConfig(ctx context.Context, cfg config.EmailConfig) (*Adapter, error) {
	adapter := NewAdapter()

	switch cfg.EmailProvider {
	case string(ProviderGoogle):
		client, err := lib_email.NewGoogleEmailClient(ctx, lib_email.GoogleEmailConfig{
			CredentialsFile: cfg.GmailCredentialsFile,
			SenderEmail:     cfg.GmailSenderEmail,
		})
		if err != nil {
			return nil, err
		}
		adapter.Register(NewEmailGoogleProvider(client))

	default:
		return nil, fmt.Errorf("email: unsupported provider %q", cfg.EmailProvider)
	}

	return adapter, nil
}
