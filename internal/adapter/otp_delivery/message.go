package otp_delivery

import (
	"errors"
	"time"

	"math-ai.com/math-ai/internal/shared/enum"
)

// Message is the channel-agnostic OTP delivery payload. The Deliverer
// concretizes it into an SMS/email/whatever payload using a template keyed
// off OtpType + Language.
//
// Code is the plaintext 6-digit value — the only place it surfaces is in the
// delivered message. The persisted ma_otps.otp_code stores its SHA-256 hash.
type Message struct {
	Identifier string        // phone (E.164) or email
	Code       string        // plaintext code to deliver (never logged)
	OtpType    enum.OtpType  // selects the template
	Language   enum.LanguageType
	ExpiresAt  time.Time
}

func (m Message) Validate() error {
	if m.Identifier == "" {
		return errors.New("otp_delivery: Identifier is required")
	}
	if m.Code == "" {
		return errors.New("otp_delivery: Code is required")
	}
	if !m.OtpType.IsValid() {
		return errors.New("otp_delivery: OtpType is invalid")
	}
	return nil
}
