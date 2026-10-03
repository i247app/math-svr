package auth

import (
	"math-ai.com/math-ai/internal/application/dto/user"
)

type LoginReq struct {
	OTPEnabled bool   `json:"otp_enabled"`
	LoginName  string `json:"login_name"`
	// Password is optional. When sent it must match the account's
	// password; an untrusted device still goes through OTP afterwards.
	Password string `json:"password,omitempty"`
}

// LoginRes carries one of two shapes depending on device trust:
type LoginRes struct {
	IsTrusted   bool               `json:"is_trusted"`
	RequiredOTP bool               `json:"required_otp"`       // if true, the client must complete the 2FA challenge for DeviceID, then re-issue /auth/login.
	OTPCode     string             `json:"otp_code,omitempty"` // otp_code exist if device not trusted
	ExpiresAt   string             `json:"expires_at,omitempty"`
	User        *user.UserResponse `json:"user"`
}

type LogoutReq struct {
	UID        *int64 `json:"-"`
	DeviceUUID string `json:"-"`
}

type LogoutRes struct{}
