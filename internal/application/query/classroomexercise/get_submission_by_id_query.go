package query

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
)

type GetSubmissionByIdQuery struct {
	ClassroomExerciseSubmissionID int64
}

type GetSubmissionByIdQueryHandler struct {
	repo domain.ISubmissionRepository
}

func NewGetSubmissionByIdQueryHandler(repo domain.ISubmissionRepository) *GetSubmissionByIdQueryHandler {
	return &GetSubmissionByIdQueryHandler{repo: repo}
}

func (h *GetSubmissionByIdQueryHandler) Handle(ctx context.Context, q GetSubmissionByIdQuery) (*domain.Submission, error) {
	return h.repo.FindBySubmissionId(ctx, q.ClassroomExerciseSubmissionID)
}
