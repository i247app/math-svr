package err_svc

import (
	"errors"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

var (
	ErrGradeNotFound            = errors.New("grade not found")
	ErrMissingGradeID           = errors.New("grade ID is required")
	ErrGradeMissingLabel        = errors.New("grade label is required")
	ErrGradeMissingDescripton   = errors.New("grade description is required")
	ErrGradeMissingDisplayOrder = errors.New("grade display order is required")
	ErrGradeAlreadyExists       = errors.New("grade already exists")
)

// NewGradeAlreadyExistsError creates a dynamic error for duplicate grade
// This will display as "Grade 'Grade 1' already exists" (localized)
func NewGradeAlreadyExistsError(label string) error {
	return &DynamicError{
		StatusCode: status.GRADE_ALREADY_EXISTS,
		Args:       map[string]interface{}{"label": label},
		BaseError:  ErrGradeAlreadyExists,
	}
}

// NewGradeNotFoundError creates a dynamic error for grade not found
// This will display as "Grade 'Grade 1' not found" (localized)
func NewGradeNotFoundError(label string) error {
	return &DynamicError{
		StatusCode: status.GRADE_NOT_FOUND,
		Args:       map[string]interface{}{"label": label},
		BaseError:  ErrGradeNotFound,
	}
}
