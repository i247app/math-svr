package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// GetExamStatsQuery reads a child's journeys. ExamType narrows to one
// type, Status to one lifecycle state (ACTIVE for "where am I now",
// COMPLETE / CANCEL for history); both nil returns everything, newest
// first inside each type.
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

func (h *GetExamStatsQueryHandler) Handle(ctx context.Context, q GetExamStatsQuery) ([]*exam.UserExam, error) {
	rows, err := h.statsRepo.ListByUserProfile(ctx, q.UserID, q.ProfileID, exam.ListJourneysFilter{
		ExamType: q.ExamType,
		Status:   q.Status,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return rows, nil
}
