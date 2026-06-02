package school

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrDescriptionTooLong = errors.New("description too long")
	ErrDistrictTooLong    = errors.New("district too long")
	ErrImageKeyTooLong    = errors.New("image_key too long")
	ErrNameCannotBeBlank  = errors.New("name cannot be blank")
	ErrNameRequired       = errors.New("name is required")
	ErrNameTooLong        = errors.New("name too long")
	ErrNoteTooLong        = errors.New("note too long")
	ErrProvinceTooLong    = errors.New("province too long")
	ErrSchoolNotFound     = errors.New("school not found")
	ErrSchoolIDRequired   = errors.New("school_id is required")
)
