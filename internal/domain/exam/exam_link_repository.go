package exam

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// AttemptResult is everything a submit writes back onto one attempt row.
// It exists so the repository has one named argument instead of five
// positional ints that are trivial to transpose.
//
// TotalQuestions counts ANSWERED questions; SkippedNumber counts the ones
// served but left blank. ScorePercentage is correct/answered, matching
// the counting rule the whole exam model is built on.
//
// EsessId is the journey the sitting folded into. Submit is the one
// moment every sitting's journey is known for certain, so it is written
// here; nil leaves whatever the hand-out recorded.
type AttemptResult struct {
	TotalQuestions  int
	CorrectNumber   int
	SkippedNumber   int
	ScorePercentage int
	SubmittedDt     time.Time
	EsessId         *int64
}

// ListAttemptsFilter narrows the attempt history. ProfileID is required —
// history is always read for one child. Status filters on
// enum.ElinkStatusType, which is how the "still open" list is built.
// EsessID narrows to one journey — the PRACTICE rounds of journey X.
type ListAttemptsFilter struct {
	ProfileID int64
	ExamType  *string
	Status    *string
	EsessID   *int64
}

// ProgressPoint is a lightweight read projection of one COMPLETED attempt
// for the learning-progress chart. It omits the JSON blobs entirely — the
// chart never needs them. ScorePercentage is non-null by construction
// (the query filters it out otherwise).
type ProgressPoint struct {
	ElinkId         int64
	ExamId          int64
	ExamType        string
	Grade           int
	ScorePercentage int64
	CorrectNumber   *int64
	TotalQuestions  *int64
	CompletedDt     time.Time
}

// ProgressPointsParams drives ListProgressPoints. From/To bound the
// submission time (nil = open). CompletedBefore fetches the prior window
// for period-over-period comparison. Limit is pre-clamped by the caller.
// EsessID narrows to one journey's sittings.
type ProgressPointsParams struct {
	ProfileID       int64
	ExamType        *string
	EsessID         *int64
	From            *time.Time
	To              *time.Time
	CompletedBefore *time.Time
	Limit           int64
}

// IExamLinkRepository owns individual attempts.
//
// MarkSubmitted is deliberately narrow rather than a general Update: an
// attempt has exactly one state transition in its life (IN_PROGRESS →
// SUBMITTED), and that transition writes the result columns in the same
// statement.
type IExamLinkRepository interface {
	FindByElinkId(ctx context.Context, elinkId int64) (*ExamLink, error)
	ListAttempts(ctx context.Context, filter ListAttemptsFilter, page, limit int64) ([]*ExamLink, *pagination.Pagination, error)
	// ListByElinkIds hydrates a set of attempts at once — the
	// journey view needs every sitting that fed a journey, and fetching
	// them one by one would turn a twenty-exam journey into twenty reads.
	ListByElinkIds(ctx context.Context, elinkIds []int64) ([]*ExamLink, error)
	// ListInProgressByProfile returns every sitting of a child that is
	// still IN_PROGRESS, oldest first. It backs the journey list, which
	// attaches each one to its journey by esess_id.
	ListInProgressByProfile(ctx context.Context, profileId int64) ([]*ExamLink, error)
	// CountHandedOutSince counts the sittings a child was handed since a
	// moment — submitted or not, because an abandoned exam still cost a
	// model call. It backs the guest daily ceiling.
	CountHandedOutSince(ctx context.Context, profileId int64, since time.Time) (int64, error)
	// ReassignOwnerByProfile re-points every sitting of one child at
	// another account. The child does not change — only who owns them —
	// so the rows are addressed by profile_id.
	ReassignOwnerByProfile(ctx context.Context, profileId int64, newUid int64) error
	// FindLatestSubmittedByEsessId returns the most recently submitted
	// sitting of a journey, whatever its type — the base a PRACTICE round
	// is drawn from. (nil, nil) when nothing has been submitted yet.
	FindLatestSubmittedByEsessId(ctx context.Context, esessId int64) (*ExamLink, error)
	// FindInProgressInJourney returns the child's newest IN_PROGRESS
	// sitting of one type in one journey, or (nil, nil). A journey holds at
	// most one per type — generate hands it back instead of opening a
	// second — but rows from before that rule may hold several.
	FindInProgressInJourney(ctx context.Context, profileId, esessId int64, examType string) (*ExamLink, error)
	// ListRecentSubmittedByProfileGrade returns a child's latest SUBMITTED
	// sittings at one grade, newest first. Their answered questions feed
	// the "do not repeat these" list the next generation at that grade is
	// prompted with; an open sitting has nothing answered yet.
	ListRecentSubmittedByProfileGrade(ctx context.Context, profileId int64, grade int, limit int) ([]*ExamLink, error)
	Create(ctx context.Context, a *ExamLink) (*ExamLink, error)
	MarkSubmitted(ctx context.Context, elinkId int64, result AttemptResult) error
	ListProgressPoints(ctx context.Context, params ProgressPointsParams) ([]*ProgressPoint, error)
}
