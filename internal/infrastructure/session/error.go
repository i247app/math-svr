package session

import "errors"

// Plain sentinel errors carried as the debug cause inside a MathError.
var (
	ErrSessionNotValid        = errors.New("session not valid")
	ErrUidNotFoundFromSession = errors.New("uid not found from session")
)
