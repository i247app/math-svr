package command

import "errors"

// Module-scoped sentinel errors for the semester command package.
var (
	ErrSemesterNotFound             = errors.New("semester not found")
	ErrSemesterNotFoundAfterInsert  = errors.New("semester not found after insert")
	ErrSemesterNotFoundAfterUpdate  = errors.New("semester not found after update")
	ErrTranslationLanguageRequired  = errors.New("translation language is required")
	ErrDuplicateTranslationLanguage = errors.New("duplicate translation language in payload")
)
