package otp_delivery

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/provider/sms"
)

// SmsDeliverer wraps the SMS provider. nil-safe: if the SMS provider is
// disabled in this deploy, Deliver returns a typed error that the OTP module
// surfaces as OTP_NO_DELIVERY_CHANNEL.
type SmsDeliverer struct {
	provider *sms.Registry
}

func NewSmsDeliverer(provider *sms.Registry) *SmsDeliverer {
	return &SmsDeliverer{provider: provider}
}

func (d *SmsDeliverer) Name() ChannelName { return ChannelSMS }

func (d *SmsDeliverer) Deliver(ctx context.Context, msg Message) error {
	if d.provider == nil {
		return errors.New("otp_delivery: sms provider is not configured")
	}

	tpl := renderTemplate(msg.OtpType, msg.Language, msg.Code, msg.ExpiresAt)
	_, err := d.provider.Send(ctx, sms.Message{
		To:   msg.Identifier,
		Body: tpl.BodyText,
	})
	return err
}
