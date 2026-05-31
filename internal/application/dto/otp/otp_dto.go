package otp

import (
	"math-ai.com/math-ai/internal/application/dto/user"
)

// SendOtpReq is the public payload for POST /otps/send and /otps/resend.
// device_uuid + device_name are optional and only meaningful for LOGIN2FA —
// they get persisted onto the OTP row so the verify caller can re-attach the
// trust decision to the right device.
//
// channel is optional. Empty (or "AUTO") = pick by identifier shape.
type SendOtpReq struct {
	OtpType    string `json:"otp_type"`
	Identifier string `json:"identifier"`
	UserID     *int64 `json:"user_id,omitempty"`
	// DeviceUUID *string    `json:"device_uuid,omitempty"`
	// DeviceName *string    `json:"device_name,omitempty"`
	// Channel    string     `json:"channel,omitempty"` // "", "SMS", "EMAIL"
	// Language   string     `json:"language,omitempty"`
}

// SendOtpRes never echoes the code. Clients use otp_id + expires_at to drive
// the verify request and the countdown timer.
type SendOtpRes struct {
	ExpiresAt string             `json:"expires_at"`
	Channel   string             `json:"-"`
	OTPCode   string             `json:"otp_code"`
	OtpType   string             `json:"otp_type"`
	User      *user.UserResponse `json:"user,omitempty"`
}

type VerifyOtpReq struct {
	OtpType    string `json:"otp_type"`
	Identifier string `json:"identifier"`
	OtpCode    string `json:"otp_code"`
}

type VerifyOtpRes struct {
	Verified bool               `json:"verified"`
	OtpType  string             `json:"otp_type"`
	User     *user.UserResponse `json:"user,omitempty"`
}

type RevokeOtpReq struct {
	OtpType    string `json:"otp_type"`
	Identifier string `json:"identifier"`
}

type RevokeOtpRes struct{}
