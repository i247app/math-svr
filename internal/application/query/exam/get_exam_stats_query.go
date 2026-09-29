package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// GetExamStatsQuery reads a child's journeys. ExamType narrows to one
// JOURNEY type, Status to one lifecycle state (ACTIVE for "where am I
// now", COMPLETE / CANCEL for history); both nil returns everything,
// newest first inside each type.
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
	ExamType  *string
	Status    *string
}

// ExamStatsResult is the journeys plus the question sets the unfinished
// sittings were drawn from, keyed by exam_id, so the caller can render
// each paper without a read per sitting.
type ExamStatsResult struct {
	Journeys  []exam.JourneyStats
	ExamPools map[int64]*exam.ExamPool
}

type GetExamStatsQueryHandler struct {
	statsRepo    exam.IExamSessionRepository
	attemptRepo  exam.IExamLinkRepository
	examPoolRepo exam.IExamPoolRepository
}

func NewGetExamStatsQueryHandler(
	statsRepo exam.IExamSessionRepository,
	attemptRepo exam.IExamLinkRepository,
	examPoolRepo exam.IExamPoolRepository,
) *GetExamStatsQueryHandler {
	return &GetExamStatsQueryHandler{statsRepo: statsRepo, attemptRepo: attemptRepo, examPoolRepo: examPoolRepo}
}

func (h *GetExamStatsQueryHandler) Handle(ctx context.Context, q GetExamStatsQuery) (*ExamStatsResult, error) {
	// Any journey — ASSESSMENT or GRADE — may carry a PRACTICE row, so the
	// type is never narrowed in SQL: that would drop the very PRACTICE rows
	// the journey needs nested. Every type is fetched and nestPractice
	// filters. (PRACTICE itself is refused by the validator — not a journey.)
	filter := exam.ListJourneysFilter{Status: q.Status}
	wantType := ""
	if q.ExamType != nil {
		wantType = *q.ExamType
	}

	rows, err := h.statsRepo.ListByUserProfile(ctx, q.UID, q.ProfileID, filter)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	journeys := nestPractice(rows, wantType)
	if len(journeys) == 0 {
		return &ExamStatsResult{Journeys: journeys, ExamPools: map[int64]*exam.ExamPool{}}, nil
	}

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
	return &ExamStatsResult{Journeys: journeys, ExamPools: examPools}, nil
}

// nestPractice folds each PRACTICE row into the journey of the same id —
// ASSESSMENT or GRADE — and drops it from the top level. wantType, when set, keeps
// only journeys of that type — the practice rows still nest, since they
// were fetched for exactly that.
func nestPractice(rows []*exam.ExamSession, wantType string) []exam.JourneyStats {
	practice := make(map[int64]*exam.ExamSession)
	var journeys []*exam.ExamSession
	for _, r := range rows {
		if r.ReqExamType() == string(enum.ExamTypePractice) {
			practice[r.EsessId()] = r
			continue
		}
		if wantType != "" && r.ReqExamType() != wantType {
			continue
		}
		journeys = append(journeys, r)
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
