package auth

import (
	"math-ai.com/math-ai/internal/application/dto/user"
)

type LoginReq struct {
	OTPEnabled bool   `json:"otp_enabled"`
	Phone      string `json:"phone"`
}

// LoginRes carries one of two shapes depending on device trust:
type LoginRes struct {
	IsTrusted   bool               `json:"is_trusted"`
	RequiredOTP bool               `json:"required_otp"`       // if true, the client must complete the 2FA challenge for DeviceID, then re-issue /auth/login.
	OTPCode     string             `json:"otp_code,omitempty"` // otp_code exist if device not trusted
	ExpiresAt   string             `json:"expires_at,omitempty"`
	User        *user.UserResponse `json:"user"`
}

type LoginWithOTPRes struct {
	ExpiresAt  string             `json:"expires_at"`
	OTPCode    string             `json:"otp_code,omitempty"`
	OtpType    string             `json:"otp_type,omitempty"`
	OTPEnabled bool               `json:"otp_enabled"`
	User       *user.UserResponse `json:"user"`
}

type LogoutReq struct{}

type LogoutRes struct{}
