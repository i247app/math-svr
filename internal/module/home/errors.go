package home

import "errors"

var (
	ErrProfileIDRequired      = errors.New("profile_id is required")
	ErrProfileNotFound        = errors.New("profile not found")
	ErrProfileNotOwnedByUser  = errors.New("profile does not belong to the session user")
	ErrUIDNotFoundFromSession = errors.New("uid not found from session")
	ErrUnsupportedRole        = errors.New("unsupported profile role for home layout")
)
