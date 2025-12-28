package err_svc

import (
	"errors"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

// Static errors (no dynamic arguments needed)
var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrInvalidUserID    = errors.New("invalid user ID")
	ErrUserNotFound     = errors.New("user not found")

	ErrMissingUID      = errors.New("uid is required")
	ErrMissingName     = errors.New("name is required")
	ErrMissingPhone    = errors.New("phone is required")
	ErrMissingPassword = errors.New("password is required")
	ErrInvalidPhone    = errors.New("invalid phone format")
	ErrMissingEmail    = errors.New("email is required")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidRole     = errors.New("invalid role")
	ErrInvalidStatus   = errors.New("invalid status")

	// Deprecated: Use ErrEmailAlreadyExistsWithValue() for localized messages with email value
	ErrEmailAlreadyExists = errors.New("email already exists")
	// Deprecated: Use ErrPhoneAlreadyExistsWithValue() for localized messages with phone value
	ErrPhoneAlreadyExists = errors.New("phone already exists")
)

// NewEmailAlreadyExistsError creates a dynamic error for duplicate email
// This will display as "Email john@example.com already exists" (localized)
func NewEmailAlreadyExistsError(email string) error {
	return &DynamicError{
		StatusCode: status.USER_EMAIL_ALREADY_EXISTS,
		Args:       map[string]interface{}{"email": email},
		BaseError:  ErrEmailAlreadyExists,
	}
}

// NewPhoneAlreadyExistsError creates a dynamic error for duplicate phone
// This will display as "Phone number +84123456789 already exists" (localized)
func NewPhoneAlreadyExistsError(phone string) error {
	return &DynamicError{
		StatusCode: status.USER_PHONE_ALREADY_EXISTS,
		Args:       map[string]interface{}{"phone": phone},
		BaseError:  ErrPhoneAlreadyExists,
	}
}

// NewUserNotFoundError creates a dynamic error for user not found
// This will display as "User with ID abc-123 not found" (localized)
func NewUserNotFoundError(userID string) error {
	return &DynamicError{
		StatusCode: status.USER_NOT_FOUND,
		Args:       map[string]interface{}{"user_id": userID},
		BaseError:  ErrUserNotFound,
	}
}
