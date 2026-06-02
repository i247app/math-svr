package auth

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrPhoneRequired   = errors.New("phone is required")
	ErrSessionNotValid = errors.New("session is not valid")
	ErrUserNotFound    = errors.New("user not found")
)
