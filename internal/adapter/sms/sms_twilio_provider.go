package sms

import (
	"context"

	"math-ai.com/math-ai/internal/libs/twilio"
)

// TwilioProvider is the SMSProvider implementation backed by
// libs/twilio. It owns nothing besides the wired *twilio.Client and
// the boot-time fallback defaults (defaultFrom, messagingServiceSID).
type TwilioProvider struct {
	client              *twilio.Client
	defaultFrom         string
	messagingServiceSID string
}

// NewTwilioProvider builds the provider. defaultFrom and msgSvcSID are
// captured here (not read from client) because the factory may want to
// pass per-deploy overrides distinct from the client's own config.
func NewTwilioProvider(c *twilio.Client, defaultFrom, msgSvcSID string) *TwilioProvider {
	return &TwilioProvider{
		client:              c,
		defaultFrom:         defaultFrom,
		messagingServiceSID: msgSvcSID,
	}
}

func (t *TwilioProvider) Name() SMSProviderName { return ProviderTwilio }

// Send invokes Twilio Messages.Create.
//
// Sender resolution precedence (Message.From > defaultFrom >
// messagingServiceSID) is fixed here; passing none yields Twilio 21603
// at the wire, which surfaces as SMS_OP_FAILED. This is intentional —
// allowing the adapter to silently succeed without a sender would
// hide a misconfiguration.
//
// The returned error is unwrapped here; adapter.Send wraps it into a
// binbaseError via mapTwilioError so log lines and HTTP responses see a
// consistent shape.
func (t *TwilioProvider) Send(ctx context.Context, msg Message) (*SendResult, error) {
	params := twilio.CreateMessageParams{
		To:   msg.To,
		Body: msg.Body,
	}

	if msg.From != "" {
		params.From = msg.From
	}
	if t.defaultFrom != "" {
		params.From = t.defaultFrom
	}
	if t.messagingServiceSID != "" {
		params.MessagingServiceSID = t.messagingServiceSID
	}

	out, err := t.client.Messages.SendByMessagingServiceSID(ctx, params)
	if err != nil {
		return nil, err
	}
	return &SendResult{
		ProviderMessageID: out.SID,
		Status:            out.Status,
	}, nil
}
