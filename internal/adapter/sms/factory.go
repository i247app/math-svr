package sms

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/libs/twilio"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
)

// NewFromConfig builds a fully-wired *Adapter from cfg, or returns
// (nil, nil) when SMS is intentionally disabled in this deploy.
//
// Behaviour matrix:
//
//	cfg.SMSProvider == "" or "disabled" → (nil, nil); boot continues.
//	cfg.SMSProvider == "twilio"         → twilio provider, registered + defaulted.
//	cfg.SMSProvider == anything else    → binbaseError(SMS_CONFIG_INVALID).
//
// Disabled returns nil rather than an error so dev profiles without
// Twilio credentials can boot. Module services that consume the
// adapter must nil-guard, the same way they do with res.Messaging in
// local dev today.
//
// Errors from libs/twilio are translated here:
//
//	ErrInvalidConfig  → binbaseError(SMS_CONFIG_INVALID, {"reason": ...})
//	anything else     → binbaseError(SMS_CONNECT_FAILED) wrapping the cause
func NewFromConfig(ctx context.Context, cfg config.SMSConfig) (*Adapter, error) {
	switch cfg.SMSProvider {
	case "", "disabled":
		return nil, nil

	case string(ProviderTwilio):
		client, err := twilio.NewClient(ctx, twilio.Config{
			AccountSID:          cfg.TwilioAccountSID,
			AuthToken:           cfg.TwilioAuthToken,
			BaseURL:             cfg.TwilioBaseURL,
			From:                cfg.TwilioFrom,
			MessagingServiceSID: cfg.TwilioMessagingServiceSID,
			Timeout:             cfg.Timeout,
			MaxRetries:          cfg.MaxRetries,
			RetryDelay:          cfg.RetryDelay,
			RequireAtBoot:       cfg.RequireAtBoot,
		})
		if err != nil {
			if errors.Is(err, twilio.ErrInvalidConfig) {
				return nil, errs.NewError(ctx, status.SMS_CONFIG_INVALID,
					map[string]any{"reason": err.Error()}, err)
			}
			return nil, errs.NewError(ctx, status.SMS_CONNECT_FAILED, nil, err)
		}

		adapter := NewAdapter()
		adapter.Register(NewTwilioProvider(client, cfg.TwilioFrom, cfg.TwilioMessagingServiceSID))
		return adapter, nil

	default:
		return nil, errs.NewError(ctx, status.SMS_CONFIG_INVALID,
			map[string]any{"provider": cfg.SMSProvider}, nil)
	}
}
