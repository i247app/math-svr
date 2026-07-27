package command

import "errors"

// Module-scoped sentinel errors for the grade command package.
var (
	ErrGradeNotFound            = errors.New("grade not found")
	ErrGradeNotFoundAfterInsert = errors.New("grade not found after insert")
	ErrGradeNotFoundAfterUpdate = errors.New("grade not found after update")
)
