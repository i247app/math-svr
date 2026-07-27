package semester

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrDisplayOrderMustBe0 = errors.New("display_order must be >= 0")
	ErrNameCannotBeBlank   = errors.New("name cannot be blank")
	ErrNameRequired        = errors.New("name is required")
	ErrNameTooLong         = errors.New("name too long")
	ErrSemesterNotFound    = errors.New("semester not found")
	ErrSemesterIDRequired  = errors.New("semester_id is required")
)
