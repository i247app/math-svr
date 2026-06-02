package command

import "errors"

// Module-scoped sentinel errors for the user command package.
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrPhoneAlreadyExists    = errors.New("phone already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)
