package otp_delivery

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/provider/email"
)

// EmailDeliverer wraps the email provider. Behaves the same as SmsDeliverer
// when the provider is nil (disabled deploy).
type EmailDeliverer struct {
	provider *email.Registry
}

func NewEmailDeliverer(provider *email.Registry) *EmailDeliverer {
	return &EmailDeliverer{provider: provider}
}

func (d *EmailDeliverer) Name() ChannelName { return ChannelEmail }

func (d *EmailDeliverer) Deliver(ctx context.Context, msg Message) error {
	if d.provider == nil {
		return errors.New("otp_delivery: email provider is not configured")
	}

	tpl := renderTemplate(msg.OtpType, msg.Language, msg.Code, msg.ExpiresAt)
	return d.provider.Send(ctx, email.Message{
		To:       []string{msg.Identifier},
		Subject:  tpl.Subject,
		BodyText: tpl.BodyText,
	})
}
