package auth

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrPhoneRequired        = errors.New("phone is required")
	ErrLoginNameRequired    = errors.New("login name is required")
	ErrSessionNotValid      = errors.New("session is not valid")
	ErrSessionNotSecure     = errors.New("session is not signed in, login required")
	ErrUserNotFound         = errors.New("user not found")
	ErrUIDNotFoundInSession = errors.New("uid not found in session")
	ErrDeviceUUIDRequired   = errors.New("device uuid is required")
)
