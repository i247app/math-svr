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
//
// UserExamId is the journey the sitting folded into. Submit is the one
// moment every sitting's journey is known for certain, so it is written
// here; nil leaves whatever the hand-out recorded.
type AttemptResult struct {
	TotalQuestions  int
	CorrectNumber   int
	SkippedNumber   int
	ScorePercentage int
	SubmittedDt     mtime.MathTime
	UserExamId      *int64
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
// UserExamID narrows to one journey — the PRACTICE rounds of journey X.
type ListAttemptsFilter struct {
	ProfileID  int64
	ExamType   *string
	Status     *string
	UserExamID *int64
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

// JourneyStats is one journey as a dashboard reads it: the row that owns
// the lifecycle, the PRACTICE row that shares its id once the child has
// practised in it (nil until then, and always nil for a journey type
// that has no practice), and every sitting of the journey still
// IN_PROGRESS — of either type, oldest first — so the child can be
// offered to finish, or pick between, what they walked away from. A
// child may be handed a new exam while one is open, so this is a list.
type JourneyStats struct {
	Journey    *UserExam
	Practice   *UserExam
	InProgress []*UserAiExam
}

// ProgressPointsParams drives ListProgressPoints. From/To bound the
// submission time (nil = open). CompletedBefore fetches the prior window
// for period-over-period comparison. Limit is pre-clamped by the caller.
// UserExamID narrows to one journey's sittings.
type ProgressPointsParams struct {
	ProfileID       int64
	ExamType        *string
	UserExamID      *int64
	From            *mtime.MathTime
	To              *mtime.MathTime
	CompletedBefore *mtime.MathTime
	Limit           int64
}

// IAiExamRepository owns the shared, user-less question sets.
//
// FindReusableByExtras is the cache read. It returns (nil, nil) on a miss
// so the caller can fall through to a real generation without treating a
// miss as an error. A set the given profile has already sat — any
// ma_user_ai_exams row of theirs pointing at it, submitted or not — is
// never a candidate: a child must not meet a paper twice, and the child
// who triggered a generation holds an attempt on it from that moment.
type IAiExamRepository interface {
	FindByAiExamId(ctx context.Context, aiExamId int64) (*AiExam, error)
	FindReusableByExtras(ctx context.Context, extras string, excludeProfileId int64) (*AiExam, error)
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
	// ListInProgressByProfile returns every sitting of a child that is
	// still IN_PROGRESS, oldest first. It backs the journey list, which
	// attaches each one to its journey by user_exam_id.
	ListInProgressByProfile(ctx context.Context, profileId int64) ([]*UserAiExam, error)
	// FindLatestSubmittedByUserExamId returns the most recently submitted
	// sitting of a journey, whatever its type — the base a PRACTICE round
	// is drawn from. (nil, nil) when nothing has been submitted yet.
	FindLatestSubmittedByUserExamId(ctx context.Context, userExamId int64) (*UserAiExam, error)
	// ListRecentByProfileGrade returns a child's latest sittings at one
	// grade, newest first, whether submitted or still open — a paper
	// handed out was seen. It feeds the "do not repeat these" list the
	// next generation at that grade is prompted with.
	ListRecentByProfileGrade(ctx context.Context, profileId int64, grade int, limit int) ([]*UserAiExam, error)
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
// A journey is one user_exam_id and holds up to TWO rows in this table:
// the ASSESSMENT row (always; it owns the lifecycle and the measured
// grade) and a PRACTICE row (once the child has submitted a practice
// round; it shares the id and keeps its own totals). The row key is
// therefore the pair (user_exam_id, req_exam_type), and every by-id read
// names the type.
//
// Opening and accumulating are two explicit operations, not one upsert.
// Create is a plain INSERT: it succeeds only when no other row holds the
// (user, profile, type) slot — or the (id, type) pair — and reports
// ErrJourneyConflict otherwise. Accumulate is an UPDATE by (id, type) that
// carries the status the caller expects in its WHERE clause — ACTIVE for
// a journey's own row, COMPLETE for a PRACTICE row — so totals can never
// be folded into a row whose state moved under the caller. Either call
// must run in the same transaction as the attempt update, or a crash
// between the two leaves the totals disagreeing with the history.
//
// MarkStatus is the only way a journey ends, and Reopen its inverse; both
// address the row that owns the lifecycle and leave a PRACTICE row alone
// — that row is born COMPLETE, when its journey is, and stays so.
// MarkStatus reports ErrJourneyNotActive when nothing was open, Reopen
// ErrJourneyNotEnded when nothing was ended and ErrJourneyConflict when
// another journey already holds the open slot.
type IUserExamRepository interface {
	// FindByUserExamId reads one row of a journey.
	FindByUserExamId(ctx context.Context, userExamId int64) (*UserExam, error)
	// FindByUserExamIdAndType reads one row of a journey. (nil, nil) when
	// the journey has no row of that type yet — a journey with no PRACTICE
	// round submitted is the ordinary case, not an error.
	FindByUserExamIdAndType(ctx context.Context, userExamId int64, examType string) (*UserExam, error)
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
	// Accumulate folds delta into the row (userExamId, examType) while it
	// is in expectedStatus, and overwrites its review from e. It never
	// touches current_grade / current_level.
	Accumulate(ctx context.Context, userExamId int64, examType, expectedStatus string, e *UserExam, delta StatsDelta) error
	// SetCurrent records the grade / level the client stated at hand-out
	// on the OPEN row (userExamId, examType); a nil value leaves that
	// column untouched. ErrJourneyNotActive when the row is not open.
	SetCurrent(ctx context.Context, userExamId int64, examType string, grade, level *int) error
	MarkStatus(ctx context.Context, userExamId int64, newStatus string, endedDt mtime.MathTime) error
	Reopen(ctx context.Context, userExamId int64) error
}

// IUserExamDetailRepository owns the per-question log.
//
// CreateBatch takes the whole sitting at once: one INSERT with N value
// tuples rather than N round-trips, since every row is written in the same
// transaction anyway.
type IUserExamDetailRepository interface {
	CreateBatch(ctx context.Context, details []*UserExamDetail) error
	ListByUserAiExamId(ctx context.Context, userAiExamId int64) ([]*UserExamDetail, error)
	// ListByUserExamId returns EVERY answered question of one row of a
	// journey — its ASSESSMENT sittings or its PRACTICE sittings — in
	// sitting order then question order: the journey review screen.
	ListByUserExamId(ctx context.Context, userExamId int64, examType string) ([]*UserExamDetail, error)
	ListRecentByUserExamId(ctx context.Context, userExamId int64, examType string, limit int64) ([]*UserExamDetail, error)
}
