package exam

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/shared/pagination"
)

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
	// FindLatestJourney returns the child's journey touched last among
	// those matching filter, in any state but DELETED, or (nil, nil). It
	// uses the same "latest" as GradeLevels.Latest, so for one GRADE grade
	// the two always name the same journey.
	FindLatestJourney(ctx context.Context, filter LatestJourneyFilter) (*ExamSession, error)
	// ListLatestCompletedByLevel returns, for one grade, the most recently
	// COMPLETED GRADE journey (newest ended_dt) of every level the child
	// has completed, lowest level first — passed or not; its esess_flag
	// says which. A level completed several times appears once; a level
	// never completed not at all.
	ListLatestCompletedByLevel(ctx context.Context, uid, profileId int64, grade int) ([]*ExamSession, error)
	// FindGradeLevels reads a child's GradeLevels for one grade, over the
	// GRADE journeys of that grade.
	FindGradeLevels(ctx context.Context, uid, profileId int64, grade int) (GradeLevels, error)
	// ListByUserProfilePage and ListByUserProfileKeyset return a child's
	// journeys (never PRACTICE rows), newest first (JourneyListKey order),
	// narrowed in SQL by filter — one page by offset, or by keyset: up to
	// Limit rows right below Older (next page), right above Newer
	// (previous page, still returned newest first), or the newest.
	ListByUserProfilePage(ctx context.Context, uid, profileId int64, filter ListJourneysFilter, page, size int64) ([]*ExamSession, *pagination.Pagination, error)
	ListByUserProfileKeyset(ctx context.Context, uid, profileId int64, filter ListJourneysFilter, params JourneyKeysetParams) ([]*ExamSession, error)
	// ExistsJourneyAtOrNewer / AtOrOlder report whether any journey of the
	// filtered list sits at key or above / below it — the keyset list's
	// "is there a page on the other side" probe.
	ExistsJourneyAtOrNewer(ctx context.Context, uid, profileId int64, filter ListJourneysFilter, key JourneyListKey) (bool, error)
	ExistsJourneyAtOrOlder(ctx context.Context, uid, profileId int64, filter ListJourneysFilter, key JourneyListKey) (bool, error)
	// ListPracticeByEsessIds returns the PRACTICE rows of the given
	// journeys of one child, whatever their state — at most one per id.
	ListPracticeByEsessIds(ctx context.Context, uid, profileId int64, esessIds []int64) ([]*ExamSession, error)
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
	// SetExamTexts records the last_title and last_short_text of the exam just
	// handed out on the OPEN row (esessId, examType). Unlike SetCurrent it
	// always writes both, nil included: the columns describe the latest
	// exam, and a latest exam without a title or summary must not show
	// the previous one's. ErrJourneyNotActive when the row is not open.
	SetExamTexts(ctx context.Context, esessId int64, examType string, title, shortText *string) error
	// SetAiReview stores the AI review on the row (esessId, examType),
	// whatever its state, WITHOUT moving modify_dt: reading a review of an
	// old journey must not make it the "latest" one. ErrJourneyNotFound
	// when no such row exists.
	SetAiReview(ctx context.Context, esessId int64, examType string, short, long string) error
	// SetLastRespID stores the journey's chain pointer (nil clears it) on
	// its OWN row — never the PRACTICE row sharing the id — whatever its
	// state, WITHOUT moving modify_dt: a model turn is not work on the
	// journey. The caller holds the journey lock.
	SetLastRespID(ctx context.Context, esessId int64, respID *string) error
	// MarkStatus ends an open journey. passMark (PassMarkFor) nil leaves
	// esess_flag NULL; otherwise esess_flag is whether the row's score, as
	// it stands in the same UPDATE, reaches passMark — a journey with no
	// score has not passed.
	MarkStatus(ctx context.Context, esessId int64, newStatus string, endedDt time.Time, passMark *int) error
	// LockJourney takes a row lock on the journey's owning row (never its
	// PRACTICE row) until the transaction ends, so hand-outs on one
	// journey run one at a time. Call it FIRST in the transaction: under
	// REPEATABLE READ a plain read made before the lock would pin a
	// snapshot from before the other hand-out committed. A journey with no
	// such row locks nothing and is not an error.
	LockJourney(ctx context.Context, esessId int64) error
	Reopen(ctx context.Context, esessId int64) error
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

// ListJourneysFilter narrows a child's journey history. Status nil means
// every journey regardless of state; ExamTypes empty means every journey
// type. PRACTICE rows are never listed as journeys (see
// IExamSessionRepository.ListPracticeByEsessIds).
type ListJourneysFilter struct {
	ExamTypes []string
	Status    *string
}

// JourneyListKey is a journey's position in the journey list, which shows
// newest first: create_dt DESC, then esess_id DESC as the tie-break (a
// journey row — never PRACTICE — is unique on esess_id). "Older" than a key
// = further down the list.
type JourneyListKey struct {
	CreateDt time.Time
	EsessId  int64
}

// JourneyKeysetParams drives IExamSessionRepository.ListByUserProfileKeyset.
// At most one of Older / Newer is set.
type JourneyKeysetParams struct {
	Older *JourneyListKey // next page: the journeys right below this one
	Newer *JourneyListKey // previous page: the journeys right above this one
	Limit int64
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
	From            *time.Time
	To              *time.Time
	SubmittedBefore *time.Time
	Limit           int64
}

// LatestJourneyFilter narrows FindLatestJourney. ExamType nil means any
// journey type; Grade nil means any grade (it filters current_grade).
// PRACTICE rows are never journeys of their own and are always excluded.
type LatestJourneyFilter struct {
	Uid       int64
	ProfileId int64
	ExamType  *string
	Grade     *int
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
