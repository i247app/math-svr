package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListClassroomsQuery powers both "my classrooms" (ProfileID set) and
// owner views. The module-level service guarantees at least one of
// ProfileID / OwnerProfileID is set before this handler is called — an
// unfiltered list of every classroom is intentionally unreachable.
type ListClassroomsQuery struct {
	ProfileID       *string
	OwnerProfileID  *string
	SchoolID        *string
	ProgramID       *string
	GradeID         *string
	Search          *string
	IncludeArchived bool
	Page            int64
	Limit           int64
}

type ListClassroomsQueryHandler struct {
	classroomRepo classroom.IRepository
}

func NewListClassroomsQueryHandler(classroomRepo classroom.IRepository) *ListClassroomsQueryHandler {
	return &ListClassroomsQueryHandler{classroomRepo: classroomRepo}
}

func (h *ListClassroomsQueryHandler) Handle(ctx context.Context, q ListClassroomsQuery) ([]*classroom.Classroom, *pagination.Pagination, error) {
	return h.classroomRepo.ListClassrooms(ctx, &classroom.ListClassroomsParams{
		ProfileId:       q.ProfileID,
		OwnerProfileId:  q.OwnerProfileID,
		SchoolId:        q.SchoolID,
		ProgramId:       q.ProgramID,
		GradeId:         q.GradeID,
		Search:          q.Search,
		IncludeArchived: q.IncludeArchived,
		Page:            q.Page,
		Limit:           q.Limit,
	})
}
