package otp_delivery

import (
	"context"

	"math-ai.com/math-ai/internal/adapter/email"
	"math-ai.com/math-ai/internal/adapter/sms"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// NewFromAdapters builds the OTP delivery adapter by composing the SMS and
// email adapters that the bootstrap layer already constructed. Either may be
// nil if the corresponding deploy disabled it; in that case the channel is
// simply not registered. Boot succeeds with zero channels — callers see
// OTP_NO_DELIVERY_CHANNEL at request time, which is preferable to crashing
// on startup.
func NewFromAdapters(ctx context.Context, smsAdapter *sms.Adapter, emailAdapter *email.Adapter) *Adapter {
	log := logger.From(ctx)
	adapter := NewAdapter()

	if smsAdapter != nil {
		adapter.Register(NewSmsDeliverer(smsAdapter))
		log.Info("otp_delivery.channel_registered channel=sms")
	}
	if emailAdapter != nil {
		adapter.Register(NewEmailDeliverer(emailAdapter))
		log.Info("otp_delivery.channel_registered channel=email")
	}

	return adapter
}
