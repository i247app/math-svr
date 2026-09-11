package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// GetExamStatsQuery reads a child's lifetime records. ExamType narrows it
// to one; leaving it nil returns every type the child has ever submitted,
// which is at most three rows.
//
// A child who has never submitted has no row at all, and that is not an
// error: the caller renders an empty state rather than a failure.
type GetExamStatsQuery struct {
	UserID    int64
	ProfileID int64
	ExamType  *string
}

type GetExamStatsQueryHandler struct {
	statsRepo exam.IUserExamRepository
}

func NewGetExamStatsQueryHandler(statsRepo exam.IUserExamRepository) *GetExamStatsQueryHandler {
	return &GetExamStatsQueryHandler{statsRepo: statsRepo}
}

func (h *GetExamStatsQueryHandler) Handle(ctx context.Context, q GetExamStatsQuery) ([]*exam.UserExam, error) {
	if q.ExamType != nil && *q.ExamType != "" {
		row, err := h.statsRepo.FindByUserProfileType(ctx, q.UserID, q.ProfileID, *q.ExamType)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if row == nil {
			return nil, nil
		}
		return []*exam.UserExam{row}, nil
	}

	rows, err := h.statsRepo.ListByUserProfile(ctx, q.UserID, q.ProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return rows, nil
}
