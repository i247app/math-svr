package command

import "errors"

// Module-scoped sentinel errors for the chapter command package.
var (
	ErrChapterNotFound              = errors.New("chapter not found")
	ErrChapterNotFoundAfterInsert   = errors.New("chapter not found after insert")
	ErrChapterNotFoundAfterUpdate   = errors.New("chapter not found after update")
	ErrTranslationLanguageRequired  = errors.New("translation language is required")
	ErrDuplicateTranslationLanguage = errors.New("duplicate translation language in payload")
)
