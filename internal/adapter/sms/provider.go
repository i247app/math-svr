package sms

import "context"

// SendResult is the provider-agnostic outcome of a successful Send.
//
// ProviderMessageID is whatever identifier the upstream returns (for
// Twilio: the message SID, e.g. "SMxxxx..."). Status mirrors the
// upstream's lifecycle state at the moment of return (Twilio: queued,
// sending, sent, delivered, failed) — note the message has typically
// not been delivered yet when this returns.
type SendResult struct {
	ProviderMessageID string
	Status            string
}

// SMSProvider is the contract every concrete SMS provider implements.
// Adapter dispatches via the registered default or a named provider.
type SMSProvider interface {
	Name() SMSProviderName
	Send(ctx context.Context, msg Message) (*SendResult, error)
}
