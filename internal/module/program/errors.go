package program

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrDescriptionCannotBeBlank              = errors.New("description cannot be blank")
	ErrDescriptionRequired                   = errors.New("description is required")
	ErrDescriptionTooLong                    = errors.New("description too long")
	ErrDisplayOrderMustBe0                   = errors.New("display_order must be >= 0")
	ErrDuplicateTranslationLanguageInPayload = errors.New("duplicate translation language in payload")
	ErrLabelCannotBeBlank                    = errors.New("label cannot be blank")
	ErrLabelRequired                         = errors.New("label is required")
	ErrLabelTooLong                          = errors.New("label too long")
	ErrLanguageMustBeVnOrEn                  = errors.New("language must be 'vn' or 'en'")
	ErrProgramNotFound                       = errors.New("program not found")
	ErrProgramIDRequired                     = errors.New("program_id is required")
	ErrTranslationDescriptionRequired        = errors.New("translation description is required")
	ErrTranslationDescriptionTooLong         = errors.New("translation description too long")
	ErrTranslationNil                        = errors.New("translation is nil")
	ErrTranslationLabelRequired              = errors.New("translation label is required")
	ErrTranslationLabelTooLong               = errors.New("translation label too long")
	ErrTranslationLanguageRequired           = errors.New("translation language is required")
	ErrTranslationLanguageTooLong            = errors.New("translation language is too long")
)
