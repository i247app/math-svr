package program

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrDescriptionCannotBeBlank = errors.New("description cannot be blank")
	ErrDescriptionRequired      = errors.New("description is required")
	ErrDescriptionTooLong       = errors.New("description too long")
	ErrDisplayOrderMustBe0      = errors.New("display_order must be >= 0")
	ErrLabelCannotBeBlank       = errors.New("label cannot be blank")
	ErrLabelRequired            = errors.New("label is required")
	ErrLabelTooLong             = errors.New("label too long")
	ErrProgramNotFound          = errors.New("program not found")
	ErrProgramIDRequired        = errors.New("program_id is required")
)
