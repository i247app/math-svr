package command

import "errors"

// Module-scoped sentinel errors for the classroom-exercise command
// package. See application/command/classroom/errors.go for the
// surrounding convention.
var (
	ErrExerciseNotFound            = errors.New("classroom exercise not found")
	ErrExerciseNotFoundAfterInsert = errors.New("classroom exercise not found after insert")
	ErrExerciseAlreadyDeleted      = errors.New("classroom exercise already deleted")
)
