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
// CallerProfileID drives the visibility access filter in the repo:
// PUBLIC rows are returned to everyone; PRIVATE rows are returned only
// when the caller is the creator. Zero (unset) collapses to PUBLIC-only.
//
// The remaining filter / sort fields mirror dto.ListExercisesReq —
// optional, normalised by the validator before reaching here.
type ListClassroomExercisesQuery struct {
	ClassroomID     int64
	CallerProfileID int64

	Status           *string
	Visibility       *string
	CreatorProfileID *int64
	ProgramID        *int64
	ChapterName      *string
	LessonName       *string
	Search           *string

	SortBy    *string
	SortOrder *string

	Page  int64
	Limit int64
}

type ListClassroomExercisesQueryHandler struct {
	repo domain.IRepository
}

func NewListClassroomExercisesQueryHandler(repo domain.IRepository) *ListClassroomExercisesQueryHandler {
	return &ListClassroomExercisesQueryHandler{repo: repo}
}

func (h *ListClassroomExercisesQueryHandler) Handle(ctx context.Context, q ListClassroomExercisesQuery) ([]*domain.Exercise, *pagination.Pagination, error) {
	return h.repo.ListExercises(ctx, domain.ListExercisesParams{
		ClassroomID:      q.ClassroomID,
		CallerProfileID:  q.CallerProfileID,
		Status:           q.Status,
		Visibility:       q.Visibility,
		CreatorProfileID: q.CreatorProfileID,
		ProgramID:        q.ProgramID,
		ChapterName:      q.ChapterName,
		LessonName:       q.LessonName,
		Search:           q.Search,
		SortBy:           q.SortBy,
		SortOrder:        q.SortOrder,
		Page:             q.Page,
		Limit:            q.Limit,
	})
}
