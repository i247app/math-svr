package sms

import (
	"errors"
	"fmt"
	"regexp"
)

// Message is the provider-agnostic SMS payload accepted by Adapter.Send.
//
// To is required and must be E.164 format ("+15551234567"). Body is
// required and capped at 1600 chars (Twilio's documented limit).
//
// From is optional. When empty the provider falls back to its
// configured default (libs/twilio.Config.From) or MessagingServiceSID.
// Per-call From wins when set on both sides.
type Message struct {
	To   string
	Body string
	From string
}

// e164 enforces E.164:
//   - leading "+"
//   - country-code first digit is 1–9 (no leading zero)
//   - 8 to 15 digits total following the "+"
var e164 = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

const maxBodyLength = 1600

// Validate enforces the documented invariants on Message. Returns plain
// errors; the Adapter.Send path wraps them in a binbaseError(SMS_OP_FAILED)
// with a "reason" arg so callers can surface the cause without leaking
// the body or the recipient.
func (m Message) Validate() error {
	if m.To == "" {
		return errors.New("sms: To is required")
	}
	if !e164.MatchString(m.To) {
		return errors.New("sms: To must be E.164 (e.g. +15551234567)")
	}
	if m.Body == "" {
		return errors.New("sms: Body is required")
	}
	if len(m.Body) > maxBodyLength {
		return fmt.Errorf("sms: Body exceeds %d chars", maxBodyLength)
	}
	return nil
}
