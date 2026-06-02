package command

import "errors"

// Module-scoped sentinel errors for the program command package.
var (
	ErrProgramNotFound              = errors.New("program not found")
	ErrProgramNotFoundAfterInsert   = errors.New("program not found after insert")
	ErrProgramNotFoundAfterUpdate   = errors.New("program not found after update")
	ErrTranslationLanguageRequired  = errors.New("translation language is required")
	ErrDuplicateTranslationLanguage = errors.New("duplicate translation language in payload")
)
