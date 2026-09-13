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
// ASSESSMENT journey it belongs to (JourneyStats.Practice), so the
// dashboard sees one entry per journey — not two rows sharing an id.
// The validator refuses ExamType = PRACTICE for that reason.
//
// A child who has never submitted has no row at all, and that is not an
// error: the caller renders an empty state rather than a failure.
type GetExamStatsQuery struct {
	UserID    int64
	ProfileID int64
	ExamType  *string
	Status    *string
}

type GetExamStatsQueryHandler struct {
	statsRepo exam.IUserExamRepository
}

func NewGetExamStatsQueryHandler(statsRepo exam.IUserExamRepository) *GetExamStatsQueryHandler {
	return &GetExamStatsQueryHandler{statsRepo: statsRepo}
}

func (h *GetExamStatsQueryHandler) Handle(ctx context.Context, q GetExamStatsQuery) ([]exam.JourneyStats, error) {
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
	return nestPractice(rows, wantType), nil
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
