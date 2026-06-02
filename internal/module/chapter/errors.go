package chapter

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrChapterNotFound                       = errors.New("chapter not found")
	ErrChapterIDRequired                     = errors.New("chapter_id is required")
	ErrDescriptionCannotBeBlank              = errors.New("description cannot be blank")
	ErrDescriptionRequired                   = errors.New("description is required")
	ErrDescriptionTooLong                    = errors.New("description too long")
	ErrDisplayOrderMustBe0                   = errors.New("display_order must be >= 0")
	ErrDuplicateTranslationLanguageInPayload = errors.New("duplicate translation language in payload")
	ErrGradeIDCannotBeBlank                  = errors.New("grade_id cannot be blank")
	ErrGradeIDRequired                       = errors.New("grade_id is required")
	ErrLabelCannotBeBlank                    = errors.New("label cannot be blank")
	ErrLabelRequired                         = errors.New("label is required")
	ErrLabelTooLong                          = errors.New("label too long")
	ErrLanguageMustBeVnOrEn                  = errors.New("language must be 'vn' or 'en'")
	ErrProgramIDCannotBeBlank                = errors.New("program_id cannot be blank")
	ErrProgramIDRequired                     = errors.New("program_id is required")
	ErrSemesterIDCannotBeBlank               = errors.New("semester_id cannot be blank")
	ErrSemesterIDRequired                    = errors.New("semester_id is required")
	ErrTranslationDescriptionRequired        = errors.New("translation description is required")
	ErrTranslationDescriptionTooLong         = errors.New("translation description too long")
	ErrTranslationNil                        = errors.New("translation is nil")
	ErrTranslationLabelRequired              = errors.New("translation label is required")
	ErrTranslationLabelTooLong               = errors.New("translation label too long")
	ErrTranslationLanguageRequired           = errors.New("translation language is required")
	ErrTranslationLanguageTooLong            = errors.New("translation language is too long")
)
