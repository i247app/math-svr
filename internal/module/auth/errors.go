package auth

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrPhoneRequired        = errors.New("phone is required")
	ErrLoginNameRequired    = errors.New("login name is required")
	ErrSessionNotValid      = errors.New("session is not valid")
	ErrUserNotFound         = errors.New("user not found")
	ErrUIDNotFoundInSession = errors.New("uid not found in session")
)
