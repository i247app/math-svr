package query

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
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
// ASSESSMENT journey it belongs to (JourneyStats.Practice), so the
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
	UserID    int64
	ProfileID int64
	ExamType  *string
	Status    *string
}

// ExamStatsResult is the journeys plus the question sets the unfinished
// sittings were drawn from, keyed by ai_exam_id, so the caller can render
// each paper without a read per sitting — and the banner over those
// journeys' cumulative scores.
type ExamStatsResult struct {
	Journeys []exam.JourneyStats
	AiExams  map[int64]*exam.AiExam
	Summary  dto.ExamStatsSummary
}

type GetExamStatsQueryHandler struct {
	statsRepo   exam.IUserExamRepository
	attemptRepo exam.IUserAiExamRepository
	aiExamRepo  exam.IAiExamRepository
}

func NewGetExamStatsQueryHandler(
	statsRepo exam.IUserExamRepository,
	attemptRepo exam.IUserAiExamRepository,
	aiExamRepo exam.IAiExamRepository,
) *GetExamStatsQueryHandler {
	return &GetExamStatsQueryHandler{statsRepo: statsRepo, attemptRepo: attemptRepo, aiExamRepo: aiExamRepo}
}

func (h *GetExamStatsQueryHandler) Handle(ctx context.Context, q GetExamStatsQuery) (*ExamStatsResult, error) {
	// An ASSESSMENT read must also pull the PRACTICE rows to nest them, so
	// it fetches every type and partitions below. Any other named type
	// has no practice and is fetched alone.
	filter := exam.ListJourneysFilter{Status: q.Status}
	wantType := ""
	if q.ExamType != nil && *q.ExamType != "" {
		wantType = *q.ExamType
		if wantType != string(enum.ExamTypeAssessment) {
			filter.ExamType = q.ExamType
		}
	}

	rows, err := h.statsRepo.ListByUserProfile(ctx, q.UserID, q.ProfileID, filter)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	journeys := nestPractice(rows, wantType)
	summary := summarizeJourneys(journeys)
	if len(journeys) == 0 {
		return &ExamStatsResult{Journeys: journeys, AiExams: map[int64]*exam.AiExam{}, Summary: summary}, nil
	}

	// One read for every unfinished sitting of the child, then attach by
	// journey. A sitting whose journey is not in this page (a filtered-out
	// type or status) is simply not shown.
	open, err := h.attemptRepo.ListInProgressByProfile(ctx, q.ProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	attached := attachInProgress(journeys, open)

	aiExams, err := hydrateAiExams(ctx, h.aiExamRepo, attached)
	if err != nil {
		return nil, err
	}
	return &ExamStatsResult{Journeys: journeys, AiExams: aiExams, Summary: summary}, nil
}

// summarizeJourneys renders the banner over the journeys being returned:
// the same aggregates the progress chart shows over sittings, here over
// each journey's cumulative score. Journeys arrive newest first and the
// trend wants chronological order, so they are walked in reverse. A
// journey with no score yet is skipped, not counted as zero; nested
// PRACTICE rows are not journeys and are not counted either.
func summarizeJourneys(journeys []exam.JourneyStats) dto.ExamStatsSummary {
	points := make([]scorePoint, 0, len(journeys))
	for i := len(journeys) - 1; i >= 0; i-- {
		j := journeys[i].Journey
		if pct := j.ResScorePercentage(); pct != nil {
			points = append(points, scorePoint{ID: j.UserExamId(), ScorePct: int64(*pct)})
		}
	}
	core, hiID := summarizeScores(points, nil)
	return dto.ExamStatsSummary{ExamScoreSummary: core, HighestUserExamID: hiID}
}

// nestPractice folds PRACTICE rows into the ASSESSMENT journey of the
// same id and drops them from the top level. wantType, when set, keeps
// only journeys of that type — the practice rows still nest, since they
// were fetched for exactly that.
func nestPractice(rows []*exam.UserExam, wantType string) []exam.JourneyStats {
	practice := make(map[int64]*exam.UserExam)
	var journeys []*exam.UserExam
	for _, r := range rows {
		if r.ReqExamType() == string(enum.ExamTypePractice) {
			practice[r.UserExamId()] = r
			continue
		}
		if wantType != "" && r.ReqExamType() != wantType {
			continue
		}
		journeys = append(journeys, r)
	}

	out := make([]exam.JourneyStats, 0, len(journeys))
	for _, j := range journeys {
		js := exam.JourneyStats{Journey: j}
		if j.ReqExamType() == string(enum.ExamTypeAssessment) {
			js.Practice = practice[j.UserExamId()]
		}
		out = append(out, js)
	}
	return out
}

// attachInProgress hangs each unfinished sitting under its journey, in
// place, and returns the ones that found a home — the set whose question
// sets need hydrating. open is oldest first and that order is kept.
func attachInProgress(journeys []exam.JourneyStats, open []*exam.UserAiExam) []*exam.UserAiExam {
	byID := make(map[int64]int, len(journeys))
	for i, j := range journeys {
		byID[j.Journey.UserExamId()] = i
	}

	var attached []*exam.UserAiExam
	for _, a := range open {
		if a.UserExamId() == nil {
			continue
		}
		i, ok := byID[*a.UserExamId()]
		if !ok {
			continue
		}
		journeys[i].InProgress = append(journeys[i].InProgress, a)
		attached = append(attached, a)
	}
	return attached
}
