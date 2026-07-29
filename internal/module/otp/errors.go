package otp

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrIdentifierRequired = errors.New("identifier is required")
	ErrOTPCodeRequired    = errors.New("otp_code is required")
	ErrOTPTypeInvalid     = errors.New("otp_type is invalid")
	ErrOTPTypeRequired    = errors.New("otp_type is required")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrTargetDeviceRequiresLogin2FA = errors.New("target_device_id is only supported for LOGIN_2FA")
)
