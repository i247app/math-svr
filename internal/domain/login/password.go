package login

import (
	"errors"
	"unicode/utf8"
)

const (
	PasswordMinChars = 8
	// PasswordMaxBytes is bcrypt's input limit: it refuses anything longer,
	// so a longer password could be accepted here and then fail to hash.
	PasswordMaxBytes = 72
)

var ErrPasswordLength = errors.New("password must be 8-72 characters (at most 72 bytes)")

// ValidatePassword checks the rules a new password must meet.
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < PasswordMinChars || len(password) > PasswordMaxBytes {
		return ErrPasswordLength
	}
	return nil
}
