package command

import "errors"

// Module-scoped sentinel errors for the auth command package. The
// "X is required" messages live here too (despite being validator-style)
// because they are the base error inside MathError, used in logs/debug.
var (
	ErrTokenNotFoundOrRevoked = errors.New("token not found or already revoked")
	ErrPhoneRequired          = errors.New("phone is required")
	ErrDeviceUUIDRequired     = errors.New("device_uuid is required")
	ErrIPAddressRequired      = errors.New("ip_address is required")
	ErrDevicePushTokenRequired = errors.New("device_push_token is required")
)
