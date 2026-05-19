package auth

import (
	"math-ai.com/math-ai/internal/application/dto/user"
)

type LoginReq struct {
	Phone string `json:"phone"`
}

// LoginRes carries one of two shapes depending on device trust:
//
//   - TwoFactorRequired=true → User is nil; the client must complete the 2FA
//     challenge for DeviceID, then re-issue /auth/login.
//   - TwoFactorRequired=false → User and DeviceID are populated; the session
//     is established.
type LoginRes struct {
	TwoFactorRequired bool               `json:"2fa_required,omitempty"`
	User              *user.UserResponse `json:"user"`
}

type LoginWithOTPRes struct {
	User    *user.UserResponse `json:"user"`
	OTPCode string             `json:"otp_code,omitempty"`
}

type LogoutReq struct{}

type LogoutRes struct{}
