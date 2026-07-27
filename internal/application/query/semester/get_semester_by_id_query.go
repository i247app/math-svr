package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/semester"
)

// GetSemesterByIdQuery returns the semester row.
type GetSemesterByIdQuery struct {
	SemesterID int64
}

type GetSemesterByIdQueryHandler struct {
	semesterRepo semester.IRepository
}

func NewGetSemesterByIdQueryHandler(semesterRepo semester.IRepository) *GetSemesterByIdQueryHandler {
	return &GetSemesterByIdQueryHandler{
		semesterRepo: semesterRepo,
	}
}

func (h *GetSemesterByIdQueryHandler) Handle(ctx context.Context, q GetSemesterByIdQuery) (*semester.Semester, error) {
	return h.semesterRepo.FindBySemesterId(ctx, q.SemesterID)
}
