package notification

import (
	"context"
	"errors"

	domain "math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/pkg/firebase"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

// NewFromConfig builds a fully-wired *Registry from cfg, or returns
// (nil, nil) when the notification provider is intentionally disabled in this
// deploy.
//
// Behaviour matrix:
//
//	cfg.Provider == "" or "disabled" → (nil, nil); boot continues.
//	cfg.Provider == "firebase"       → Firebase provider, registered + defaulted.
//	cfg.Provider == anything else    → MathError(NOTIFICATION_CONFIG_INVALID).
//
// Disabled returns nil rather than an error so dev profiles without Firebase
// credentials can boot. Module services that consume the registry must
// nil-guard, the same way they do with res.SMSProvider in local dev.
func NewFromConfig(ctx context.Context, cfg config.NotificationConfig) (*Registry, error) {
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
				return nil, errs.NewError(ctx, domain.NotificationConfigInvalid,
					map[string]any{"reason": err.Error()}, err)
			}
			return nil, errs.NewError(ctx, domain.NotificationConnectFailed, nil, err)
		}

		reg := NewRegistry()
		reg.Register(NewFirebaseProvider(client))
		return reg, nil

	default:
		return nil, errs.NewError(ctx, domain.NotificationConfigInvalid,
			map[string]any{"provider": cfg.Provider}, nil)
	}
}
