package command

import "errors"

// Module-scoped sentinel errors for the classroom-exercise-submission
// command package. See application/command/classroom/errors.go for the
// surrounding convention.
var (
	ErrSubmissionNotFound            = errors.New("submission not found")
	ErrSubmissionNotFoundAfterInsert = errors.New("submission not found after insert")
	ErrSubmissionAlreadyDeleted      = errors.New("submission already deleted")
	ErrSeqReturnedZeroID             = errors.New("seq returned zero id")
)
