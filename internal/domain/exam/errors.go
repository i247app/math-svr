package exam

import "errors"

// ErrAttemptNotInProgress reports that a submit hit an attempt that was
// not IN_PROGRESS — already submitted, deleted, or gone.
//
// It exists because the check cannot live only in the command: two submits
// racing on the same attempt both read IN_PROGRESS before either writes.
// The repository's UPDATE carries the state in its WHERE clause, so the
// loser updates zero rows and gets this error back, and the application
// layer maps it to EXAM_ALREADY_SUBMITTED.
var ErrAttemptNotInProgress = errors.New("exam: attempt is not in progress")
