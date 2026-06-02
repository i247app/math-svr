package device

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrDeviceNotFound   = errors.New("device not found")
	ErrDeviceIDRequired = errors.New("device_id is required")
	ErrUserIDRequired   = errors.New("user_id is required")
)
