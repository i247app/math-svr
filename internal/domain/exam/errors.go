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

// ErrJourneyNotActive reports that a mark hit a journey that was not
// ACTIVE — already ended, deleted, or gone. Same shape and same reason as
// ErrAttemptNotInProgress: the repository's UPDATE carries the expected
// state in its WHERE clause, so two marks racing on one journey cannot
// both succeed.
var ErrJourneyNotActive = errors.New("exam: journey is not active")

// ErrJourneyConflict reports that opening a journey collided with a row
// that already holds the (user, profile, type) slot — in practice, another
// submit opened the journey a moment earlier. The caller re-reads the open
// journey and folds into it instead.
//
// It is a distinct error rather than a silent merge so that a collision
// on ANY unique key surfaces here. An INSERT ... ON DUPLICATE KEY UPDATE
// would have quietly redirected the write into whichever row happened to
// collide — which, when the schema once drifted, turned out to be an
// already-ended journey.
var ErrJourneyConflict = errors.New("exam: journey slot already taken")
