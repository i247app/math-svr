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
	Create(ctx context.Context, a *UserAiExam) (*UserAiExam, error)
	MarkSubmitted(ctx context.Context, userAiExamId int64, result AttemptResult) error
	ListProgressPoints(ctx context.Context, params ProgressPointsParams) ([]*ProgressPoint, error)
}

// IUserExamRepository owns the lifetime totals.
//
// Upsert accumulates delta onto the (user, profile, type) row, creating it
// when absent. The placement fields carried on e (review, grade, level)
// overwrite; the counters in delta add. It must be called inside the same
// transaction as the attempt update, or a crash between the two leaves the
// totals disagreeing with the attempt history.
type IUserExamRepository interface {
	FindByUserProfileType(ctx context.Context, userId, profileId int64, examType string) (*UserExam, error)
	// ListByUserProfile returns every exam type's lifetime row for one
	// child — at most three rows, and what a statistics screen shows.
	ListByUserProfile(ctx context.Context, userId, profileId int64) ([]*UserExam, error)
	Upsert(ctx context.Context, e *UserExam, delta StatsDelta) error
}

// IUserExamDetailRepository owns the per-question log.
//
// CreateBatch takes the whole sitting at once: one INSERT with N value
// tuples rather than N round-trips, since every row is written in the same
// transaction anyway.
type IUserExamDetailRepository interface {
	CreateBatch(ctx context.Context, details []*UserExamDetail) error
	ListByUserAiExamId(ctx context.Context, userAiExamId int64) ([]*UserExamDetail, error)
	ListRecentByUserExamId(ctx context.Context, userExamId int64, limit int64) ([]*UserExamDetail, error)
}
