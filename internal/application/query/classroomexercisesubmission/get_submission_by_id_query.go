package query

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/classroomexercisesubmission"
)

type GetSubmissionByIdQuery struct {
	ClassroomExerciseSubmissionID int64
}

type GetSubmissionByIdQueryHandler struct {
	repo domain.IRepository
}

func NewGetSubmissionByIdQueryHandler(repo domain.IRepository) *GetSubmissionByIdQueryHandler {
	return &GetSubmissionByIdQueryHandler{repo: repo}
}

func (h *GetSubmissionByIdQueryHandler) Handle(ctx context.Context, q GetSubmissionByIdQuery) (*domain.Submission, error) {
	return h.repo.FindBySubmissionId(ctx, q.ClassroomExerciseSubmissionID)
}
