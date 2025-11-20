package err_svc

import "errors"

var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrInvalidUserID    = errors.New("invalid user ID")
	ErrUserNotFound     = errors.New("user not found")

	ErrMissingFirstName   = errors.New("first name is required")
	ErrMissingLastName    = errors.New("last name is required")
	ErrMissingPhone       = errors.New("phone is required")
	ErrMissingPassword    = errors.New("password is required")
	ErrInvalidPhone       = errors.New("invalid phone format")
	ErrPhoneAlreadyExists = errors.New("phone already exists")
	ErrMissingEmail       = errors.New("email is required")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidStatus      = errors.New("invalid status")
)
