package otp_delivery

import (
	"context"
	"errors"
	"strings"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/libs/twilio"
)

// Adapter is the entry point the OTP application layer uses to dispatch
// codes. It owns a small set of channel deliverers and routes a Message
// either to an explicit channel (SendVia) or to the channel that matches the
// identifier shape (Send).
type Adapter struct {
	deliverers map[ChannelName]Deliverer
}

func NewAdapter() *Adapter {
	return &Adapter{deliverers: make(map[ChannelName]Deliverer)}
}

// Register adds a deliverer. Calling Register with a nil-wrapping deliverer
// is allowed — it stays registered so SendVia can return a useful error
// instead of "channel not configured". The caller is responsible for not
// registering a deliverer when the underlying adapter is disabled.
func (a *Adapter) Register(d Deliverer) {
	if a.deliverers == nil {
		a.deliverers = make(map[ChannelName]Deliverer)
	}
	a.deliverers[d.Name()] = d
}

// Send picks a channel based on the identifier shape: "@" → email, else SMS.
// Returns an error if the inferred channel isn't registered.
func (a *Adapter) Send(ctx context.Context, msg Message) error {
	channel := inferChannel(msg.Identifier)
	return a.SendVia(ctx, channel, msg)
}

// SendVia dispatches msg through the named channel.
func (a *Adapter) SendVia(ctx context.Context, name ChannelName, msg Message) error {
	if err := msg.Validate(); err != nil {
		return err
	}

	d, ok := a.deliverers[name]
	if !ok {
		return errors.New("otp_delivery: channel " + string(name) + " is not registered")
	}

	log := logger.From(ctx)
	if err := d.Deliver(ctx, msg); err != nil {
		// Log discipline: never include the plaintext code, never the
		// raw identifier when it's a phone number.
		log.Warnf("otp_delivery.failed channel=%s to=%s type=%s err=%v",
			name, maskIdentifier(msg.Identifier), msg.OtpType, err)
		return err
	}
	log.Infof("otp_delivery.sent channel=%s to=%s type=%s", name, maskIdentifier(msg.Identifier), msg.OtpType)
	return nil
}

// HasChannel reports whether the named channel is registered. The OTP module
// uses this to decide whether to even attempt a send.
func (a *Adapter) HasChannel(name ChannelName) bool {
	_, ok := a.deliverers[name]
	return ok
}

// inferChannel guesses email vs sms from the identifier shape. "@" is the
// only signal — phone identifiers must always be E.164 ("+...").
func inferChannel(identifier string) ChannelName {
	if strings.Contains(identifier, "@") {
		return ChannelEmail
	}
	return ChannelSMS
}

func maskIdentifier(id string) string {
	if strings.Contains(id, "@") {
		// Mask the local part of the email: "ab***@example.com"
		at := strings.Index(id, "@")
		if at <= 2 {
			return "***" + id[at:]
		}
		return id[:2] + "***" + id[at:]
	}
	return twilio.MaskPhone(id)
}
