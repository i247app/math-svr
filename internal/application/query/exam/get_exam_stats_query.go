package query

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// GetExamStatsQuery reads a child's journeys. ExamTypes narrows to the
// listed JOURNEY types (empty = all), Status to one lifecycle state
// (ACTIVE for "where am I now", COMPLETE / CANCEL for history); both
// filters run in SQL. Newest first, one page at a time — by offset or by
// cursor (Paging, already validated).
//
// PRACTICE rows are never returned on their own. Each rides inside the
// journey it belongs to — ASSESSMENT or GRADE — (JourneyStats.Practice), so the
// dashboard sees one entry per journey — not two rows sharing an id.
// The validator refuses ExamType = PRACTICE for that reason.
//
// Each journey also carries its unfinished sittings. A journey is opened
// the moment an exam is handed out, so a child who generated one and
// walked away has a journey row here, with the sitting under it, and can
// be offered to finish it — that is the whole reason the row is opened
// so early.
//
// A child who has never generated has no row at all, and that is not an
// error: the caller renders an empty state rather than a failure.
type GetExamStatsQuery struct {
	UID       int64
	ProfileID int64
	ExamTypes []string
	Status    *string
	Paging    pagination.Request
}

// ExamStatsResult is the journeys plus the question sets the unfinished
// sittings were drawn from, keyed by exam_id, so the caller can render
// each paper without a read per sitting. Exactly one of Pagination
// (OFFSET) / Cursor (CURSOR) is set.
type ExamStatsResult struct {
	Journeys   []exam.JourneyStats
	ExamPools  map[int64]*exam.ExamPool
	Pagination *pagination.Pagination
	Cursor     *pagination.CursorPagination
}

// journeyCursor is what a /exams/sessions/list cursor holds: a journey's
// position in the list (exam.JourneyListKey). create_dt is kept to the
// microsecond (DATETIME(6), RFC 3339 in JSON), so it compares equal to the
// stored value.
type journeyCursor struct {
	CreateDt time.Time `json:"create_dt"`
	EsessId  int64     `json:"esess_id"`
}

func (c journeyCursor) key() exam.JourneyListKey {
	return exam.JourneyListKey{CreateDt: c.CreateDt, EsessId: c.EsessId}
}

type GetExamStatsQueryHandler struct {
	examSessionRepo exam.IExamSessionRepository
	attemptRepo     exam.IExamLinkRepository
	examPoolRepo    exam.IExamPoolRepository
}

func NewGetExamStatsQueryHandler(
	examSessionRepo exam.IExamSessionRepository,
	attemptRepo exam.IExamLinkRepository,
	examPoolRepo exam.IExamPoolRepository,
) *GetExamStatsQueryHandler {
	return &GetExamStatsQueryHandler{examSessionRepo: examSessionRepo, attemptRepo: attemptRepo, examPoolRepo: examPoolRepo}
}

func (h *GetExamStatsQueryHandler) Handle(ctx context.Context, q GetExamStatsQuery) (*ExamStatsResult, error) {
	// Type and status narrow the journeys in SQL. The PRACTICE rows are
	// then read for exactly the journeys returned — by id, whatever their
	// own state: a PRACTICE row stays COMPLETE when its journey is
	// reopened, and it still belongs to that journey. (PRACTICE itself is
	// refused by the validator — not a journey.)
	filter := exam.ListJourneysFilter{ExamTypes: q.ExamTypes, Status: q.Status}
	result := &ExamStatsResult{Journeys: []exam.JourneyStats{}, ExamPools: map[int64]*exam.ExamPool{}}
	var rows []*exam.ExamSession
	var err error
	if q.Paging.IsCursor() {
		rows, result.Cursor, err = pagination.KeysetPage(ctx, h.journeyKeyset(q.UID, q.ProfileID, filter), q.Paging.Next, q.Paging.Previous, q.Paging.Size)
	} else {
		rows, result.Pagination, err = h.examSessionRepo.ListByUserProfilePage(ctx, q.UID, q.ProfileID, filter, q.Paging.Page, q.Paging.Size)
	}
	if err != nil {
		if _, ok := errs.IsMathError(err); ok {
			return nil, err // a bad cursor — already PAGINATION_INVALID_CURSOR
		}
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(rows) == 0 {
		return result, nil
	}

	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.EsessId())
	}
	practice, err := h.examSessionRepo.ListPracticeByEsessIds(ctx, q.UID, q.ProfileID, ids)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	journeys := nestPractice(rows, practice)

	// One read for every unfinished sitting of the child, then attach by
	// journey. A sitting whose journey is not in this page (a filtered-out
	// type or status) is simply not shown.
	open, err := h.attemptRepo.ListInProgressByProfile(ctx, q.ProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	attached := attachInProgress(journeys, open)

	examPools, err := hydrateExamPools(ctx, h.examPoolRepo, attached)
	if err != nil {
		return nil, err
	}
	result.Journeys, result.ExamPools = journeys, examPools
	return result, nil
}

// journeyKeyset is the child's filtered journey list as a keyset source.
func (h *GetExamStatsQueryHandler) journeyKeyset(uid, profileId int64, filter exam.ListJourneysFilter) pagination.Keyset[*exam.ExamSession, journeyCursor] {
	return pagination.Keyset[*exam.ExamSession, journeyCursor]{
		Read: func(ctx context.Context, after, before *journeyCursor, limit int64) ([]*exam.ExamSession, error) {
			params := exam.JourneyKeysetParams{Limit: limit}
			if after != nil {
				k := after.key()
				params.Older = &k
			}
			if before != nil {
				k := before.key()
				params.Newer = &k
			}
			return h.examSessionRepo.ListByUserProfileKeyset(ctx, uid, profileId, filter, params)
		},
		// Newest first: "before" a journey in the list means newer.
		ExistsAtOrBefore: func(ctx context.Context, c journeyCursor) (bool, error) {
			return h.examSessionRepo.ExistsJourneyAtOrNewer(ctx, uid, profileId, filter, c.key())
		},
		ExistsAtOrAfter: func(ctx context.Context, c journeyCursor) (bool, error) {
			return h.examSessionRepo.ExistsJourneyAtOrOlder(ctx, uid, profileId, filter, c.key())
		},
		Key: func(j *exam.ExamSession) journeyCursor {
			return journeyCursor{CreateDt: j.CreateDt().ToTime().UTC(), EsessId: j.EsessId()}
		},
		Valid: func(c journeyCursor) bool { return c.EsessId > 0 && !c.CreateDt.IsZero() },
	}
}

// nestPractice hangs each PRACTICE row under the journey of the same id —
// ASSESSMENT or GRADE — keeping the journeys' order.
func nestPractice(journeys, practiceRows []*exam.ExamSession) []exam.JourneyStats {
	practice := make(map[int64]*exam.ExamSession, len(practiceRows))
	for _, p := range practiceRows {
		practice[p.EsessId()] = p
	}

	out := make([]exam.JourneyStats, 0, len(journeys))
	for _, j := range journeys {
		out = append(out, exam.JourneyStats{Journey: j, Practice: practice[j.EsessId()]})
	}
	return out
}

// attachInProgress hangs each unfinished sitting under its journey, in
// place, and returns the ones that found a home — the set whose question
// sets need hydrating. open is oldest first and that order is kept.
func attachInProgress(journeys []exam.JourneyStats, open []*exam.ExamLink) []*exam.ExamLink {
	byID := make(map[int64]int, len(journeys))
	for i, j := range journeys {
		byID[j.Journey.EsessId()] = i
	}

	var attached []*exam.ExamLink
	for _, a := range open {
		if a.EsessId() == nil {
			continue
		}
		i, ok := byID[*a.EsessId()]
		if !ok {
			continue
		}
		journeys[i].InProgress = append(journeys[i].InProgress, a)
		attached = append(attached, a)
	}
	return attached
}
