package err_svc

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

// DynamicError is a custom error type that supports localized messages with dynamic arguments
// It allows passing runtime values (like email addresses, user names, etc.) to error messages
type DynamicError struct {
	// StatusCode is the application-specific status code
	StatusCode status.Code

	// Args contains dynamic arguments to be interpolated into the error message
	// Use named keys for better clarity and maintainability
	Args map[string]interface{}

	// BaseError is the underlying error (optional)
	BaseError error
}

// Error implements the error interface
// Returns a basic English message for logging purposes
func (e *DynamicError) Error() string {
	if e.BaseError != nil {
		if len(e.Args) > 0 {
			return fmt.Sprintf("%v (args: %v)", e.BaseError.Error(), e.Args)
		}
		return e.BaseError.Error()
	}

	if len(e.Args) > 0 {
		return fmt.Sprintf("Error with status %v (args: %v)", e.StatusCode, e.Args)
	}

	return fmt.Sprintf("Error with status %v", e.StatusCode)
}

// GetStatusCode returns the status code
func (e *DynamicError) GetStatusCode() status.Code {
	return e.StatusCode
}

// GetArgs returns the dynamic arguments
func (e *DynamicError) GetArgs() map[string]interface{} {
	if e.Args == nil {
		return make(map[string]interface{})
	}
	return e.Args
}

// Unwrap returns the base error (for error wrapping)
func (e *DynamicError) Unwrap() error {
	return e.BaseError
}

// NewDynamicError creates a new DynamicError with status code and dynamic arguments
func NewDynamicError(statusCode status.Code, args map[string]interface{}) *DynamicError {
	return &DynamicError{
		StatusCode: statusCode,
		Args:       args,
		BaseError:  nil,
	}
}

// NewDynamicErrorWithBase creates a DynamicError wrapping an existing error
func NewDynamicErrorWithBase(statusCode status.Code, baseError error, args map[string]interface{}) *DynamicError {
	return &DynamicError{
		StatusCode: statusCode,
		Args:       args,
		BaseError:  baseError,
	}
}

// Helper functions for common error scenarios

// ErrEmailAlreadyExistsWithValue creates a dynamic error for duplicate email with the actual email value
func ErrEmailAlreadyExistsWithValue(email string) *DynamicError {
	return NewDynamicError(status.USER_EMAIL_ALREADY_EXISTS, map[string]interface{}{
		"email": email,
	})
}

// ErrPhoneAlreadyExistsWithValue creates a dynamic error for duplicate phone with the actual phone value
func ErrPhoneAlreadyExistsWithValue(phone string) *DynamicError {
	return NewDynamicError(status.USER_PHONE_ALREADY_EXISTS, map[string]interface{}{
		"phone": phone,
	})
}

// ErrUserNotFoundWithID creates a dynamic error for user not found with the user ID
func ErrUserNotFoundWithID(userID string) *DynamicError {
	return NewDynamicError(status.USER_NOT_FOUND, map[string]interface{}{
		"user_id": userID,
	})
}

// ErrGradeAlreadyExistsWithLabel creates a dynamic error for duplicate grade with the label
func ErrGradeAlreadyExistsWithLabel(label string) *DynamicError {
	return NewDynamicError(status.GRADE_ALREADY_EXISTS, map[string]interface{}{
		"label": label,
	})
}

// ErrSemesterAlreadyExistsWithName creates a dynamic error for duplicate semester with the name
func ErrSemesterAlreadyExistsWithName(name string) *DynamicError {
	return NewDynamicError(status.SEMESTER_ALREADY_EXISTS, map[string]interface{}{
		"name": name,
	})
}

// IsDynamicError checks if an error is a DynamicError
func IsDynamicError(err error) (*DynamicError, bool) {
	if err == nil {
		return nil, false
	}

	if dynErr, ok := err.(*DynamicError); ok {
		return dynErr, true
	}

	return nil, false
}
