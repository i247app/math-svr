package game

import "errors"

// ErrSessionNotInProgress reports that a submit hit a round that was not
// IN_PROGRESS — already submitted, deleted, or gone. The repository's UPDATE
// carries the state in its WHERE clause, so of two submits racing on one
// round the loser gets this error instead of grading it twice.
var ErrSessionNotInProgress = errors.New("game: session is not in progress")
