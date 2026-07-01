package notification

import (
	"context"
	"errors"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/libs/firebase"
)

// NewFromConfig builds a fully-wired *Adapter from cfg, or returns
// (nil, nil) when the notification adapter is intentionally disabled in this
// deploy.
//
// Behaviour matrix:
//
//	cfg.Provider == "" or "disabled" → (nil, nil); boot continues.
//	cfg.Provider == "firebase"       → Firebase provider, registered + defaulted.
//	cfg.Provider == anything else    → MathError(NOTIFICATION_CONFIG_INVALID).
//
// Disabled returns nil rather than an error so dev profiles without Firebase
// credentials can boot. Module services that consume the adapter must
// nil-guard, the same way they do with res.SMSProvider in local dev.
func NewFromConfig(ctx context.Context, cfg config.NotificationConfig) (*Adapter, error) {
	switch cfg.Provider {
	case "", "disabled":
		return nil, nil

	case string(ProviderFirebase):
		client, err := firebase.NewClient(ctx, firebase.Config{
			CredentialsFile: cfg.FirebaseCredentialsFile,
			ProjectID:       cfg.FirebaseProjectID,
		})
		if err != nil {
			if errors.Is(err, firebase.ErrInvalidConfig) {
				return nil, errs.NewError(ctx, status.NOTIFICATION_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			return nil, errs.NewError(ctx, status.NOTIFICATION_CONNECT_FAILED, nil, err)
		}

		adapter := NewAdapter()
		adapter.Register(NewFirebaseProvider(client))
		return adapter, nil

	default:
		return nil, errs.NewError(ctx, status.NOTIFICATION_CONFIG_INVALID,
			map[string]any{"provider": cfg.Provider}, nil)
	}
}
