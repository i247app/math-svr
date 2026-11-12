package otp_delivery

import (
	"context"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/provider/email"
	"math-ai.com/math-ai/internal/provider/sms"
)

// NewFromProviders builds the OTP delivery provider by composing the SMS and
// email providers that the bootstrap layer already constructed. Either may be
// nil if the corresponding deploy disabled it; in that case the channel is
// simply not registered. Boot succeeds with zero channels — callers see
// OTP_NO_DELIVERY_CHANNEL at request time, which is preferable to crashing
// on startup.
func NewFromProviders(ctx context.Context, smsProvider *sms.Registry, emailProvider *email.Registry) *Registry {
	log := logger.From(ctx)
	reg := NewRegistry()

	if smsProvider != nil {
		reg.Register(NewSmsDeliverer(smsProvider))
		log.Info("otp_delivery.channel_registered channel=sms")
	}
	if emailProvider != nil {
		reg.Register(NewEmailDeliverer(emailProvider))
		log.Info("otp_delivery.channel_registered channel=email")
	}

	return reg
}
