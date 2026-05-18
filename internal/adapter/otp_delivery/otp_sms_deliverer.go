package otp_delivery

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/adapter/sms"
)

// SmsDeliverer wraps the SMS adapter. nil-safe: if the SMS adapter is
// disabled in this deploy, Deliver returns a typed error that the OTP module
// surfaces as OTP_NO_DELIVERY_CHANNEL.
type SmsDeliverer struct {
	adapter *sms.Adapter
}

func NewSmsDeliverer(adapter *sms.Adapter) *SmsDeliverer {
	return &SmsDeliverer{adapter: adapter}
}

func (d *SmsDeliverer) Name() ChannelName { return ChannelSMS }

func (d *SmsDeliverer) Deliver(ctx context.Context, msg Message) error {
	if d.adapter == nil {
		return errors.New("otp_delivery: sms adapter is not configured")
	}

	tpl := renderTemplate(msg.OtpType, msg.Language, msg.Code, msg.ExpiresAt)
	_, err := d.adapter.Send(ctx, sms.Message{
		To:   msg.Identifier,
		Body: tpl.BodyText,
	})
	return err
}
