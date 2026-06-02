package command

import "errors"

// Module-scoped sentinel errors for the otp command package.
var (
	ErrInvalidOtpType       = errors.New("invalid otp type")
	ErrIdentifierRequired   = errors.New("identifier is required")
	ErrOtpCodeRequired      = errors.New("otp code is required")
	ErrNoPendingOtp         = errors.New("no pending otp")
	ErrOtpExpired           = errors.New("otp expired")
	ErrOtpAttemptCapExceeded = errors.New("attempt cap exceeded")
	ErrOtpCodeMismatch      = errors.New("code mismatch")
	ErrOtpSendWindowReached = errors.New("send-window cap reached")
	ErrOtpUnknownChannel    = errors.New("unknown channel")
	ErrOtpChannelNotRegistered = errors.New("channel not registered")
)
