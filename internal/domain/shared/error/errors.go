package errs

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/shared/enum"

	"math-ai.com/math-ai/internal/domain/shared/status"
)

type MathError struct {
	// StatusCode is the application-specific status code
	Status *status.MStatus

	// BaseError is the underlying error (optional)
	BaseError error
}

// Error implements the error interface
// Returns a basic English message for logging purposes
func (e *MathError) Error() string {
	if e.BaseError != nil {
		return e.BaseError.Error()
	}

	isSuccess := e.Status.Code() == status.SUCCESS || e.Status.Code() == status.CREATED
	if isSuccess {
		return fmt.Sprintf("Success with status %v", e.Status.Code())
	}

	return fmt.Sprintf("Error with status %v", e.Status.Code())
}

// NewDynamicError creates a new DynamicError with status code and dynamic arguments
func NewError(ctx context.Context, statusCode status.StatusCode, args map[string]any, baseError error) *MathError {
	// language := ctx.Value("language").(enum.LanguageType)
	language := enum.LanguageTypeEnglish

	message := status.GetMessage(language, statusCode, args)

	return &MathError{
		Status:    status.NewMStatus(statusCode, message),
		BaseError: baseError,
	}
}

// GetStatusCode returns the status code
func (e *MathError) GetStatusCode() status.StatusCode {
	return e.Status.Code()
}

// GetStatusMessage returns the status message
func (e *MathError) GetStatusMessage() status.StatusMessage {
	return e.Status.Message()
}

// Unwrap returns the base error (for error wrapping)
func (e *MathError) Unwrap() error {
	return e.BaseError
}

// IsbinbaseError checks if an error is a DynamicError
func IsMathError(err error) (*MathError, bool) {
	if err == nil {
		return nil, false
	}

	if binbaseErr, ok := err.(*MathError); ok {
		return binbaseErr, true
	}

	return nil, false
}

func NewInvalidParamError(ctx context.Context, baseError error) *MathError {
	return NewError(ctx, status.FAIL, nil, baseError)
}

func NewUnauthorizedError(ctx context.Context, baseError error) *MathError {
	return NewError(ctx, status.UNAUTHORIZED, nil, baseError)
}
