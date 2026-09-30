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
// EsessId is the journey the sitting folded into. Submit is the one
// moment every sitting's journey is known for certain, so it is written
// here; nil leaves whatever the hand-out recorded.
type AttemptResult struct {
	TotalQuestions  int
	CorrectNumber   int
	SkippedNumber   int
	ScorePercentage int
	SubmittedDt     mtime.MathTime
	EsessId         *int64
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
	Journey    *ExamSession
	Practice   *ExamSession
	InProgress []*ExamLink
}

// ProgressPointsParams drives ListProgressPoints. From/To bound the
// submission time (nil = open). CompletedBefore fetches the prior window
// for period-over-period comparison. Limit is pre-clamped by the caller.
// EsessID narrows to one journey's sittings.
type ProgressPointsParams struct {
	ProfileID       int64
	ExamType        *string
	EsessID         *int64
	From            *mtime.MathTime
	To              *mtime.MathTime
	CompletedBefore *mtime.MathTime
	Limit           int64
}

// IExamPoolRepository owns the shared, user-less question sets.
//
// FindReusableByExtras is the cache read. It returns (nil, nil) on a miss
// so the caller can fall through to a real generation without treating a
// miss as an error. A set the given profile has already sat — any
// ma_exam_links row of theirs pointing at it, submitted or not — is
// never a candidate: a child must not meet a paper twice, and the child
// who triggered a generation holds an attempt on it from that moment.
type IExamPoolRepository interface {
	FindByExamId(ctx context.Context, examId int64) (*ExamPool, error)
	FindReusableByExtras(ctx context.Context, extras string, excludeProfileId int64) (*ExamPool, error)
	// CountByExtras reports how many question sets already sit under one
	// cache tag. The caller uses it to decide whether the pool is deep
	// enough to serve from — see the variant threshold in module/exam.
	CountByExtras(ctx context.Context, extras string) (int64, error)
	// ListByExamIds hydrates many question sets at once. A history list
	// renders one title per attempt, and fetching them one by one would
	// turn a twenty-row screen into twenty-one queries.
	ListByExamIds(ctx context.Context, examIds []int64) ([]*ExamPool, error)
	Create(ctx context.Context, e *ExamPool) (*ExamPool, error)
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
	CountHandedOutSince(ctx context.Context, profileId int64, since mtime.MathTime) (int64, error)
	// ReassignOwnerByProfile re-points every sitting of one child at
	// another account. The child does not change — only who owns them —
	// so the rows are addressed by profile_id.
	ReassignOwnerByProfile(ctx context.Context, profileId int64, newUid int64) error
	// FindLatestSubmittedByEsessId returns the most recently submitted
	// sitting of a journey, whatever its type — the base a PRACTICE round
	// is drawn from. (nil, nil) when nothing has been submitted yet.
	FindLatestSubmittedByEsessId(ctx context.Context, esessId int64) (*ExamLink, error)
	// ListRecentByProfileGrade returns a child's latest sittings at one
	// grade, newest first, whether submitted or still open — a paper
	// handed out was seen. It feeds the "do not repeat these" list the
	// next generation at that grade is prompted with.
	ListRecentByProfileGrade(ctx context.Context, profileId int64, grade int, limit int) ([]*ExamLink, error)
	Create(ctx context.Context, a *ExamLink) (*ExamLink, error)
	MarkSubmitted(ctx context.Context, elinkId int64, result AttemptResult) error
	ListProgressPoints(ctx context.Context, params ProgressPointsParams) ([]*ProgressPoint, error)
}

// ListJourneysFilter narrows a child's journey history. Status nil means
// every journey regardless of state; ExamType nil means every type.
type ListJourneysFilter struct {
	ExamType *string
	Status   *string
}

// JourneyProgressParams drives IExamSessionRepository.ListProgressPoints,
// the journey-level counterpart of ProgressPointsParams. A journey's
// point in time is its last_submitted_dt — the moment its cumulative
// score last moved — so From/To and SubmittedBefore all bound that
// column. ExamType nil means every journey type except PRACTICE, whose
// rows are not journeys of their own. Limit is pre-clamped by the caller.
type JourneyProgressParams struct {
	UID             int64
	ProfileID       int64
	ExamType        *string
	From            *mtime.MathTime
	To              *mtime.MathTime
	SubmittedBefore *mtime.MathTime
	Limit           int64
}

// GradeLevels is where a child stands on the GRADE level ladder of one
// grade. Each GRADE journey is one "lock" at one level (its
// current_level); the client moves the child up after a lock is
// COMPLETE and down after a round that fell short.
//
// Both read the GRADE journeys of that grade in any state but DELETED.
// Latest is the level of the journey touched last — where the child is
// now, including after a step down. Max is the highest level the child
// has ever reached: a journey exists at a level only once the client has
// moved the child up to it, so an ACTIVE journey counts as much as a
// finished one. COMPLETE is not a pass mark (a journey is completed with
// any score), so it is no filter here. Each is nil when no journey
// qualifies.
type GradeLevels struct {
	Latest *int
	Max    *int
}

// IExamSessionRepository owns journeys.
//
// A journey is one esess_id and holds up to TWO rows in this table:
// the owning row — ASSESSMENT or GRADE, whichever the journey was opened
// as (always; it owns the lifecycle and the measured grade) and a PRACTICE row (once the child has submitted a practice
// round; it shares the id and keeps its own totals). The row key is
// therefore the pair (esess_id, req_exam_type). A by-id read either
// names the type or uses FindByEsessId for the owning row — never a
// hard-coded ASSESSMENT, which refuses every GRADE journey.
//
// Opening and accumulating are two explicit operations, not one upsert.
// Create is a plain INSERT: it succeeds only when no other row holds the
// journey's slot (JourneyKey) — or the (id, type) pair — and reports
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
type IExamSessionRepository interface {
	// FindByEsessId reads the row that owns the journey's lifecycle (the
	// ASSESSMENT or GRADE row), never the PRACTICE row sharing its id.
	FindByEsessId(ctx context.Context, esessId int64) (*ExamSession, error)
	// FindByEsessIdAndType reads one row of a journey. (nil, nil) when
	// the journey has no row of that type yet — a journey with no PRACTICE
	// round submitted is the ordinary case, not an error.
	FindByEsessIdAndType(ctx context.Context, esessId int64, examType string) (*ExamSession, error)
	// FindActiveJourney returns the open journey of one slot, or (nil, nil)
	// when the slot is empty. A GRADE key without a grade may match several
	// open journeys (one per grade); the most recently touched one wins.
	FindActiveJourney(ctx context.Context, key JourneyKey) (*ExamSession, error)
	// FindLatestCompletedByUserProfileType returns the most recently
	// COMPLETED journey of that type — what a new journey inherits its
	// starting grade from. CANCELLED journeys are skipped on purpose.
	FindLatestCompletedByUserProfileType(ctx context.Context, uid, profileId int64, examType string) (*ExamSession, error)
	// FindGradeLevels reads a child's GradeLevels for one grade, over the
	// GRADE journeys of that grade.
	FindGradeLevels(ctx context.Context, uid, profileId int64, grade int) (GradeLevels, error)
	// ListByUserProfile returns a child's journeys, newest first within
	// each exam type.
	ListByUserProfile(ctx context.Context, uid, profileId int64, filter ListJourneysFilter) ([]*ExamSession, error)
	// ListProgressPoints returns a child's scored journeys, newest
	// submission first, capped at params.Limit — the journey-level series
	// behind the progress chart. A journey nothing was ever submitted in
	// has no score and is left out.
	ListProgressPoints(ctx context.Context, params JourneyProgressParams) ([]*ExamSession, error)
	// ReassignOwnerByProfile re-points every journey of one child at
	// another account. uk_active_journey keys on (uid, profile_id, type[,
	// grade]), and the profile moves with its journeys, so an open journey stays
	// open and cannot collide with one the receiving account already has.
	ReassignOwnerByProfile(ctx context.Context, profileId int64, newUid int64) error
	// Create opens a journey with delta as its first totals.
	Create(ctx context.Context, e *ExamSession, delta StatsDelta) error
	// Accumulate folds delta into the row (esessId, examType) while it
	// is in expectedStatus, and overwrites its review from e. It never
	// touches current_grade / current_level.
	Accumulate(ctx context.Context, esessId int64, examType, expectedStatus string, e *ExamSession, delta StatsDelta) error
	// SetCurrent records the grade / level the client stated at hand-out
	// on the OPEN row (esessId, examType); a nil value leaves that
	// column untouched. ErrJourneyNotActive when the row is not open.
	SetCurrent(ctx context.Context, esessId int64, examType string, grade, level *int) error
	MarkStatus(ctx context.Context, esessId int64, newStatus string, endedDt mtime.MathTime) error
	Reopen(ctx context.Context, esessId int64) error
}

// IExamSessionLineRepository owns the per-question log.
//
// CreateBatch takes the whole sitting at once: one INSERT with N value
// tuples rather than N round-trips, since every row is written in the same
// transaction anyway.
type IExamSessionLineRepository interface {
	CreateBatch(ctx context.Context, details []*ExamSessionLine) error
	ListByElinkId(ctx context.Context, elinkId int64) ([]*ExamSessionLine, error)
	// ListByEsessId returns EVERY answered question of one row of a
	// journey — its ASSESSMENT sittings or its PRACTICE sittings — in
	// sitting order then question order: the journey review screen.
	ListByEsessId(ctx context.Context, esessId int64, examType string) ([]*ExamSessionLine, error)
	ListRecentByEsessId(ctx context.Context, esessId int64, examType string, limit int64) ([]*ExamSessionLine, error)
}
