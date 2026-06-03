package query

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListSubmissionsQuery is the unified read shape for every list path:
// my-list (ProfileID set), by-exercise (ClassroomExerciseID set),
// by-classroom (ClassroomID set). Multiple filters AND together, so a
// caller can scope by (classroom, exercise) at once when needed.
type ListSubmissionsQuery struct {
	ClassroomID         int64
	ClassroomExerciseID int64
	ProfileID           int64

	Status *string

	SortBy    *string
	SortOrder *string

	Page  int64
	Limit int64
}

type ListSubmissionsQueryHandler struct {
	repo domain.ISubmissionRepository
}

func NewListSubmissionsQueryHandler(repo domain.ISubmissionRepository) *ListSubmissionsQueryHandler {
	return &ListSubmissionsQueryHandler{repo: repo}
}

func (h *ListSubmissionsQueryHandler) Handle(ctx context.Context, q ListSubmissionsQuery) ([]*domain.Submission, *pagination.Pagination, error) {
	return h.repo.ListSubmissions(ctx, domain.ListSubmissionsParams{
		ClassroomID:         q.ClassroomID,
		ClassroomExerciseID: q.ClassroomExerciseID,
		ProfileID:           q.ProfileID,
		Status:              q.Status,
		SortBy:              q.SortBy,
		SortOrder:           q.SortOrder,
		Page:                q.Page,
		Limit:               q.Limit,
	})
}
