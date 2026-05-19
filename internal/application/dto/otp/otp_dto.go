package otp

import (
	"time"

	"github.com/google/uuid"
)

// SendOtpReq is the public payload for POST /otps/send and /otps/resend.
// device_uuid + device_name are optional and only meaningful for LOGIN2FA —
// they get persisted onto the OTP row so the verify caller can re-attach the
// trust decision to the right device.
//
// channel is optional. Empty (or "AUTO") = pick by identifier shape.
type SendOtpReq struct {
	OtpType    string     `json:"otp_type"`
	Identifier string     `json:"identifier"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	// DeviceUUID *string    `json:"device_uuid,omitempty"`
	// DeviceName *string    `json:"device_name,omitempty"`
	// Channel    string     `json:"channel,omitempty"` // "", "SMS", "EMAIL"
	// Language   string     `json:"language,omitempty"`
}

// SendOtpRes never echoes the code. Clients use otp_id + expires_at to drive
// the verify request and the countdown timer.
type SendOtpRes struct {
	OtpID     uuid.UUID `json:"otp_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Channel   string    `json:"channel"`
	OTPCode   string    `json:"otp_code"`
}

type VerifyOtpReq struct {
	OtpType    string `json:"otp_type"`
	Identifier string `json:"identifier"`
	OtpCode    string `json:"otp_code"`
}

type VerifyOtpRes struct {
	OtpID    uuid.UUID `json:"otp_id"`
	Verified bool      `json:"verified"`
}

type RevokeOtpReq struct {
	OtpType    string `json:"otp_type"`
	Identifier string `json:"identifier"`
}

type RevokeOtpRes struct{}
