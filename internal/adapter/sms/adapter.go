package sms

import (
	"context"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/libs/twilio"

	"math-ai.com/math-ai/internal/domain/shared/status"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
)

// Adapter is the entry point callers use to dispatch SMS. It holds one
// or more registered providers and a designated default used by Send.
//
// The Adapter is safe for concurrent use after construction; provider
// registration is expected to happen at boot inside NewFromConfig.
type Adapter struct {
	providers   map[SMSProviderName]SMSProvider
	defaultName SMSProviderName
}

func NewAdapter() *Adapter {
	return &Adapter{providers: make(map[SMSProviderName]SMSProvider)}
}

// Register adds a provider to the adapter. The first provider
// registered becomes the default; callers may override that with
// SetDefault.
func (a *Adapter) Register(provider SMSProvider) {
	if a.providers == nil {
		a.providers = make(map[SMSProviderName]SMSProvider)
	}
	a.providers[provider.Name()] = provider
	if a.defaultName == "" {
		a.defaultName = provider.Name()
	}
}

// SetDefault selects which registered provider Send dispatches to. It
// returns an error if the provider name has not been registered.
func (a *Adapter) SetDefault(name SMSProviderName) error {
	if _, ok := a.providers[name]; !ok {
		return fmt.Errorf("sms: provider %q is not registered", name)
	}
	a.defaultName = name
	return nil
}

// Send dispatches msg through the adapter's default provider after
// running Message.Validate.
func (a *Adapter) Send(ctx context.Context, msg Message) (*SendResult, error) {
	if a.defaultName == "" {
		return nil, errors.New("sms: no default provider configured")
	}
	return a.SendVia(ctx, a.defaultName, msg)
}

// SendVia dispatches msg through the provider registered under name.
//
// Validation happens here so failures are caught uniformly regardless
// of which routing entry point is used; the underlying provider is
// never called on an invalid payload, which keeps Twilio billing clean.
func (a *Adapter) SendVia(ctx context.Context, name SMSProviderName, msg Message) (*SendResult, error) {
	log := logger.From(ctx)

	if err := msg.Validate(); err != nil {
		return nil, errs.NewError(ctx, status.SMS_OP_FAILED, map[string]any{
			"reason":    err.Error(),
			"to_masked": twilio.MaskPhone(msg.To),
		}, err)
	}

	provider, ok := a.providers[name]
	if !ok {
		return nil, fmt.Errorf("sms: provider %q is not registered", name)
	}

	res, err := provider.Send(ctx, msg)
	if err != nil {
		kerr := mapTwilioError(ctx, msg, err)
		// Log discipline (spec §13): masked phone only; never the body,
		// never the recipient in cleartext, never any auth material.
		switch kerr.GetStatusCode() {
		case status.SMS_CONFIG_INVALID:
			log.Errorf("sms.config_error provider=%s to=%s err=%v", name, twilio.MaskPhone(msg.To), err)
		case status.SMS_SERIALIZE_FAILED:
			log.Errorf("sms.serialize_error provider=%s to=%s err=%v", name, twilio.MaskPhone(msg.To), err)
		default:
			// SMS_OP_FAILED: covers transient 5xx and validation-style 4xx.
			log.Warnf("sms.op_error provider=%s to=%s err=%v", name, twilio.MaskPhone(msg.To), err)
		}
		return nil, kerr
	}

	log.Infof("sms.sent provider=%s to=%s sid=%s status=%s", name, twilio.MaskPhone(msg.To), res.ProviderMessageID, res.Status)
	return res, nil
}
