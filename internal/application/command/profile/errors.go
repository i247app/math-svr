package command

import "errors"

// Module-scoped sentinel errors for the profile command package.
var (
	ErrProfileNotFound          = errors.New("profile not found")
	ErrSchoolNotFound           = errors.New("school not found")
	ErrProfileCodeMintExhausted = errors.New("could not mint a unique profile code")
)
