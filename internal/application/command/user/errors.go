package command

import "errors"

// Module-scoped sentinel errors for the user command package.
var (
	ErrUserNotFound             = errors.New("user not found")
	ErrPhoneAlreadyExists       = errors.New("phone already exists")
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrUsernameAlreadyExists    = errors.New("username already exists")
	ErrProfileCodeMintExhausted = errors.New("could not mint a unique profile code")
	ErrEmailNotVerified         = errors.New("email not verified via a matching REGISTER otp within the trust window")
)
