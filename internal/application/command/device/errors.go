package command

import "errors"

// Module-scoped sentinel errors for the device command package.
var (
	ErrDeviceNotFound        = errors.New("device not found")
	ErrDeviceNotOwnedByUser  = errors.New("device does not belong to user")
	ErrDeviceAlreadyVerified = errors.New("device already verified")
)
