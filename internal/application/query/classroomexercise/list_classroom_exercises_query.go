package query

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListClassroomExercisesQuery scopes the read to a single classroom.
// Status is optional — empty / nil returns every non-deleted lifecycle
// state (ACTIVE + ARCHIVED, the repo already filters DELETED).
//
// CallerProfileID drives the visibility filter in the repo: PUBLIC rows
// are returned to everyone; PRIVATE rows are returned only when the
// caller is the creator. Zero (unset) collapses to PUBLIC-only.
type ListClassroomExercisesQuery struct {
	ClassroomID     int64
	CallerProfileID int64
	Status          *string
	Page            int64
	Limit           int64
}

type ListClassroomExercisesQueryHandler struct {
	repo domain.IRepository
}

func NewListClassroomExercisesQueryHandler(repo domain.IRepository) *ListClassroomExercisesQueryHandler {
	return &ListClassroomExercisesQueryHandler{repo: repo}
}

func (h *ListClassroomExercisesQueryHandler) Handle(ctx context.Context, q ListClassroomExercisesQuery) ([]*domain.Exercise, *pagination.Pagination, error) {
	return h.repo.ListExercises(ctx, domain.ListExercisesParams{
		ClassroomID:     q.ClassroomID,
		CallerProfileID: q.CallerProfileID,
		Status:          q.Status,
		Page:            q.Page,
		Limit:           q.Limit,
	})
}
