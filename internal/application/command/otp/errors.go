package command

import "errors"

// Module-scoped sentinel errors for the otp command package.
var (
	ErrInvalidOtpType          = errors.New("invalid otp type")
	ErrIdentifierRequired      = errors.New("identifier is required")
	ErrOtpCodeRequired         = errors.New("otp code is required")
	ErrNoPendingOtp            = errors.New("no pending otp")
	ErrOtpExpired              = errors.New("otp expired")
	ErrOtpAttemptCapExceeded   = errors.New("attempt cap exceeded")
	ErrOtpCodeMismatch         = errors.New("code mismatch")
	ErrOtpSendWindowReached    = errors.New("send-window cap reached")
	ErrOtpUnknownChannel       = errors.New("unknown channel")
	ErrOtpChannelNotRegistered = errors.New("channel not registered")

	// Trusted-device push 2FA (target_device_id) sentinels.
	ErrTargetDeviceNotFound     = errors.New("target device not found")
	ErrTargetDeviceNotOwned     = errors.New("target device not owned by user")
	ErrTargetDeviceNotTrusted   = errors.New("target device is not trusted")
	ErrTargetDeviceNoPushToken  = errors.New("target device has no push token")
	ErrPushChannelNotRegistered = errors.New("push channel not registered")
	ErrPushDeliveryFailed       = errors.New("push delivery failed")
	ErrPushTokenInvalid         = errors.New("push token rejected as invalid")
)
