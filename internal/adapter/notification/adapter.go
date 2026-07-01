package notification

import (
	"context"
	"errors"
	"fmt"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// Adapter is the entry point callers use to dispatch push notifications. It
// holds one or more registered providers and a designated default used by
// Send. Safe for concurrent use after construction.
type Adapter struct {
	providers   map[NotificationProviderName]NotificationProvider
	defaultName NotificationProviderName
}

func NewAdapter() *Adapter {
	return &Adapter{providers: make(map[NotificationProviderName]NotificationProvider)}
}

// Register adds a provider to the adapter. The first provider registered
// becomes the default; callers may override that with SetDefault.
func (a *Adapter) Register(provider NotificationProvider) {
	if a.providers == nil {
		a.providers = make(map[NotificationProviderName]NotificationProvider)
	}
	a.providers[provider.Name()] = provider
	if a.defaultName == "" {
		a.defaultName = provider.Name()
	}
}

// SetDefault selects which registered provider Send dispatches to. It returns
// an error if the provider name has not been registered.
func (a *Adapter) SetDefault(name NotificationProviderName) error {
	if _, ok := a.providers[name]; !ok {
		return fmt.Errorf("notification: provider %q is not registered", name)
	}
	a.defaultName = name
	return nil
}

// Send dispatches msg through the adapter's default provider.
func (a *Adapter) Send(ctx context.Context, msg PushMessage) (*SendResult, error) {
	if a.defaultName == "" {
		return nil, errors.New("notification: no default provider configured")
	}
	return a.SendVia(ctx, a.defaultName, msg)
}

// SendVia dispatches msg through the provider registered under name.
//
// Validation happens here so an empty token set or empty payload never
// reaches the vendor. Tokens themselves are never logged (operator context
// only) — only the recipient count.
func (a *Adapter) SendVia(ctx context.Context, name NotificationProviderName, msg PushMessage) (*SendResult, error) {
	log := logger.From(ctx)

	if err := msg.Validate(); err != nil {
		return nil, errs.NewError(ctx, status.NOTIFICATION_SEND_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}

	provider, ok := a.providers[name]
	if !ok {
		return nil, fmt.Errorf("notification: provider %q is not registered", name)
	}

	res, err := provider.Send(ctx, msg)
	if err != nil {
		log.Warnf("notification.send_error provider=%s tokens=%d err=%v", name, len(msg.Tokens), err)
		if _, ok := errs.IsMathError(err); ok {
			return nil, err
		}
		return nil, errs.NewError(ctx, status.NOTIFICATION_SEND_FAILED,
			map[string]any{"provider": string(name)}, err)
	}

	log.Infof("notification.send provider=%s tokens=%d success=%d failure=%d invalid=%d",
		name, len(msg.Tokens), res.SuccessCount, res.FailureCount, len(res.InvalidTokens))
	return res, nil
}
