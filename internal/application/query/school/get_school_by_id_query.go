package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/school"
)

type GetSchoolByIdQuery struct {
	SchoolID string
}

type GetSchoolByIdQueryHandler struct {
	schoolRepo school.IRepository
}

func NewGetSchoolByIdQueryHandler(schoolRepo school.IRepository) *GetSchoolByIdQueryHandler {
	return &GetSchoolByIdQueryHandler{schoolRepo: schoolRepo}
}

func (h *GetSchoolByIdQueryHandler) Handle(ctx context.Context, q GetSchoolByIdQuery) (*school.School, error) {
	return h.schoolRepo.FindBySchoolId(ctx, q.SchoolID)
}
