package query

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
)

type GetClassroomExerciseByIdQuery struct {
	ClassroomExerciseID int64
}

type GetClassroomExerciseByIdQueryHandler struct {
	repo domain.IRepository
}

func NewGetClassroomExerciseByIdQueryHandler(repo domain.IRepository) *GetClassroomExerciseByIdQueryHandler {
	return &GetClassroomExerciseByIdQueryHandler{repo: repo}
}

func (h *GetClassroomExerciseByIdQueryHandler) Handle(ctx context.Context, q GetClassroomExerciseByIdQuery) (*domain.Exercise, error) {
	return h.repo.FindByClassroomExerciseId(ctx, q.ClassroomExerciseID)
}
