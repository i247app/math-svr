package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/grade"
)

// GetGradeByIdQuery returns the grade row.
type GetGradeByIdQuery struct {
	GradeID int64
}

type GetGradeByIdQueryHandler struct {
	gradeRepo grade.IRepository
}

func NewGetGradeByIdQueryHandler(gradeRepo grade.IRepository) *GetGradeByIdQueryHandler {
	return &GetGradeByIdQueryHandler{
		gradeRepo: gradeRepo,
	}
}

func (h *GetGradeByIdQueryHandler) Handle(ctx context.Context, q GetGradeByIdQuery) (*grade.Grade, error) {
	return h.gradeRepo.FindByGradeId(ctx, q.GradeID)
}
