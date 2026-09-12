package exam

import (
	"context"

	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// AttemptResult is everything a submit writes back onto one attempt row.
// It exists so the repository has one named argument instead of five
// positional ints that are trivial to transpose.
//
// TotalQuestions counts ANSWERED questions; SkippedNumber counts the ones
// served but left blank. ScorePercentage is correct/answered, matching
// the counting rule the whole exam model is built on.
type AttemptResult struct {
	TotalQuestions  int
	CorrectNumber   int
	SkippedNumber   int
	ScorePercentage int
	SubmittedDt     mtime.MathTime
}

// StatsDelta is the set of counters ADDED to a lifetime row on submit, as
// opposed to the placement fields (review, grade, level) which are
// overwritten. Keeping the two apart in the type is what stops a caller
// from accidentally overwriting a lifetime total with a single sitting's.
type StatsDelta struct {
	TotalQuestions int
	CorrectNumber  int
	SkippedNumber  int
}

// ListAttemptsFilter narrows the attempt history. ProfileID is required —
// history is always read for one child. Status filters on
// enum.UserAiExamStatusType, which is how the "still open" list is built.
type ListAttemptsFilter struct {
	ProfileID int64
	ExamType  *string
	Status    *string
}

// ProgressPoint is a lightweight read projection of one COMPLETED attempt
// for the learning-progress chart. It omits the JSON blobs entirely — the
// chart never needs them. ScorePercentage is non-null by construction
// (the query filters it out otherwise).
type ProgressPoint struct {
	UserAiExamId    int64
	AiExamId        int64
	ExamType        string
	Grade           int
	ScorePercentage int64
	CorrectNumber   *int64
	TotalQuestions  *int64
	CompletedDt     mtime.MathTime
}

// ProgressPointsParams drives ListProgressPoints. From/To bound the
// submission time (nil = open). CompletedBefore fetches the prior window
// for period-over-period comparison. Limit is pre-clamped by the caller.
type ProgressPointsParams struct {
	ProfileID       int64
	ExamType        *string
	From            *mtime.MathTime
	To              *mtime.MathTime
	CompletedBefore *mtime.MathTime
	Limit           int64
}

// IAiExamRepository owns the shared, user-less question sets.
//
// FindReusableByExtras is the cache read. It returns (nil, nil) on a miss
// so the caller can fall through to a real generation without treating a
// miss as an error.
type IAiExamRepository interface {
	FindByAiExamId(ctx context.Context, aiExamId int64) (*AiExam, error)
	FindReusableByExtras(ctx context.Context, extras string) (*AiExam, error)
	// CountByExtras reports how many question sets already sit under one
	// cache tag. The caller uses it to decide whether the pool is deep
	// enough to serve from — see the variant threshold in module/exam.
	CountByExtras(ctx context.Context, extras string) (int64, error)
	// ListByAiExamIds hydrates many question sets at once. A history list
	// renders one title per attempt, and fetching them one by one would
	// turn a twenty-row screen into twenty-one queries.
	ListByAiExamIds(ctx context.Context, aiExamIds []int64) ([]*AiExam, error)
	Create(ctx context.Context, e *AiExam) (*AiExam, error)
}

// IUserAiExamRepository owns individual attempts.
//
// MarkSubmitted is deliberately narrow rather than a general Update: an
// attempt has exactly one state transition in its life (IN_PROGRESS →
// SUBMITTED), and that transition writes the result columns in the same
// statement.
type IUserAiExamRepository interface {
	FindByUserAiExamId(ctx context.Context, userAiExamId int64) (*UserAiExam, error)
	ListAttempts(ctx context.Context, filter ListAttemptsFilter, page, limit int64) ([]*UserAiExam, *pagination.Pagination, error)
	// ListByUserAiExamIds hydrates a set of attempts at once — the
	// journey view needs every sitting that fed a journey, and fetching
	// them one by one would turn a twenty-exam journey into twenty reads.
	ListByUserAiExamIds(ctx context.Context, userAiExamIds []int64) ([]*UserAiExam, error)
	Create(ctx context.Context, a *UserAiExam) (*UserAiExam, error)
	MarkSubmitted(ctx context.Context, userAiExamId int64, result AttemptResult) error
	ListProgressPoints(ctx context.Context, params ProgressPointsParams) ([]*ProgressPoint, error)
}

// ListJourneysFilter narrows a child's journey history. Status nil means
// every journey regardless of state; ExamType nil means every type.
type ListJourneysFilter struct {
	ExamType *string
	Status   *string
}

// IUserExamRepository owns journeys.
//
// Opening and accumulating are two explicit operations, not one upsert.
// Create is a plain INSERT: it succeeds only when no other row holds the
// (user, profile, type) slot, and reports ErrJourneyConflict otherwise.
// Accumulate is an UPDATE by id that carries ACTIVE in its WHERE clause,
// so totals can never be folded into a journey that has ended. Either
// call must run in the same transaction as the attempt update, or a
// crash between the two leaves the totals disagreeing with the history.
//
// MarkStatus is the only way a journey ends. It too carries ACTIVE as the
// expected state and reports ErrJourneyNotActive when no row matched, so
// two marks racing on one journey cannot both "win".
type IUserExamRepository interface {
	FindByUserExamId(ctx context.Context, userExamId int64) (*UserExam, error)
	// FindActiveByUserProfileType returns the open journey, or (nil, nil)
	// when the child has none of that type right now.
	FindActiveByUserProfileType(ctx context.Context, userId, profileId int64, examType string) (*UserExam, error)
	// FindLatestCompletedByUserProfileType returns the most recently
	// COMPLETED journey of that type — what a new journey inherits its
	// starting grade from. CANCELLED journeys are skipped on purpose.
	FindLatestCompletedByUserProfileType(ctx context.Context, userId, profileId int64, examType string) (*UserExam, error)
	// ListByUserProfile returns a child's journeys, newest first within
	// each exam type.
	ListByUserProfile(ctx context.Context, userId, profileId int64, filter ListJourneysFilter) ([]*UserExam, error)
	// Create opens a journey with delta as its first totals.
	Create(ctx context.Context, e *UserExam, delta StatsDelta) error
	// Accumulate folds delta into the OPEN journey userExamId and
	// overwrites its placement fields from e.
	Accumulate(ctx context.Context, userExamId int64, e *UserExam, delta StatsDelta) error
	MarkStatus(ctx context.Context, userExamId int64, newStatus string, endedDt mtime.MathTime) error
}

// IUserExamDetailRepository owns the per-question log.
//
// CreateBatch takes the whole sitting at once: one INSERT with N value
// tuples rather than N round-trips, since every row is written in the same
// transaction anyway.
type IUserExamDetailRepository interface {
	CreateBatch(ctx context.Context, details []*UserExamDetail) error
	ListByUserAiExamId(ctx context.Context, userAiExamId int64) ([]*UserExamDetail, error)
	// ListByUserExamId returns EVERY answered question of a journey, in
	// sitting order then question order — the journey review screen.
	ListByUserExamId(ctx context.Context, userExamId int64) ([]*UserExamDetail, error)
	ListRecentByUserExamId(ctx context.Context, userExamId int64, limit int64) ([]*UserExamDetail, error)
}
