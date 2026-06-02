package semester

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrDisplayOrderMustBe0                   = errors.New("display_order must be >= 0")
	ErrDuplicateTranslationLanguageInPayload = errors.New("duplicate translation language in payload")
	ErrLanguageMustBeVnOrEn                  = errors.New("language must be 'vn' or 'en'")
	ErrNameCannotBeBlank                     = errors.New("name cannot be blank")
	ErrNameRequired                          = errors.New("name is required")
	ErrNameTooLong                           = errors.New("name too long")
	ErrSemesterNotFound                      = errors.New("semester not found")
	ErrSemesterIDRequired                    = errors.New("semester_id is required")
	ErrTranslationNil                        = errors.New("translation is nil")
	ErrTranslationLanguageRequired           = errors.New("translation language is required")
	ErrTranslationLanguageTooLong            = errors.New("translation language is too long")
	ErrTranslationNameRequired               = errors.New("translation name is required")
	ErrTranslationNameTooLong                = errors.New("translation name too long")
)
