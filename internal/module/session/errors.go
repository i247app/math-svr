package session

import "errors"

var (
	ErrTokenRequired    = errors.New("token is required")
	ErrSessionNotFound  = errors.New("no session is stored under that token")
	ErrIsSecureRequired = errors.New("is_secure is required")
	ErrSecureWithoutUID = errors.New("a session without a uid cannot be marked secure")
)
