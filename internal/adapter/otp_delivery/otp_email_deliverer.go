package otp_delivery

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/adapter/email"
)

// EmailDeliverer wraps the email adapter. Behaves the same as SmsDeliverer
// when the adapter is nil (disabled deploy).
type EmailDeliverer struct {
	adapter *email.Adapter
}

func NewEmailDeliverer(adapter *email.Adapter) *EmailDeliverer {
	return &EmailDeliverer{adapter: adapter}
}

func (d *EmailDeliverer) Name() ChannelName { return ChannelEmail }

func (d *EmailDeliverer) Deliver(ctx context.Context, msg Message) error {
	if d.adapter == nil {
		return errors.New("otp_delivery: email adapter is not configured")
	}

	tpl := renderTemplate(msg.OtpType, msg.Language, msg.Code, msg.ExpiresAt)
	return d.adapter.Send(ctx, email.Message{
		To:       []string{msg.Identifier},
		Subject:  tpl.Subject,
		BodyText: tpl.BodyText,
	})
}
