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
//
// ProgramID and ProgramIDs are unioned at the repo layer (OR semantics).
// Callers may set either or both; both blank means "no program filter".
type ListClassroomsQuery struct {
	ProfileID       *int64
	OwnerProfileID  *int64
	SchoolID        *int64
	ProgramID       *int64
	ProgramIDs      []int64
	GradeID         *int64
	Search          *string
	ClassCode       *string
	IncludeArchived bool
	Page            int64
	Limit           int64
}

type ListClassroomsQueryHandler struct {
	classroomRepo        classroom.IRepository
	classroomProgramRepo classroom.IClassroomProgramRepository
}

func NewListClassroomsQueryHandler(
	classroomRepo classroom.IRepository,
	classroomProgramRepo classroom.IClassroomProgramRepository,
) *ListClassroomsQueryHandler {
	return &ListClassroomsQueryHandler{
		classroomRepo:        classroomRepo,
		classroomProgramRepo: classroomProgramRepo,
	}
}

// Handle fetches the classroom page and then hydrates each row's
// in-memory ProgramIds slice via a single batched junction read. The
// batch keeps the page-level cost at one extra query regardless of
// pagination size, avoiding the per-row N+1 a naive loop would create.
func (h *ListClassroomsQueryHandler) Handle(ctx context.Context, q ListClassroomsQuery) ([]*classroom.Classroom, *pagination.Pagination, error) {
	classrooms, pg, err := h.classroomRepo.ListClassrooms(ctx, &classroom.ListClassroomsParams{
		ProfileId:       q.ProfileID,
		OwnerProfileId:  q.OwnerProfileID,
		SchoolId:        q.SchoolID,
		ProgramId:       q.ProgramID,
		ProgramIds:      q.ProgramIDs,
		GradeId:         q.GradeID,
		Search:          q.Search,
		ClassCode:       q.ClassCode,
		IncludeArchived: q.IncludeArchived,
		Page:            q.Page,
		Limit:           q.Limit,
	})
	if err != nil {
		return nil, nil, err
	}
	if len(classrooms) == 0 {
		return classrooms, pg, nil
	}

	ids := make([]int64, len(classrooms))
	for i, c := range classrooms {
		ids[i] = c.ClassroomId()
	}
	links, err := h.classroomProgramRepo.ListProgramIdsByClassroomIds(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	for _, c := range classrooms {
		ps, ok := links[c.ClassroomId()]
		if !ok {
			ps = []int64{}
		}
		c.SetProgramIds(ps)
	}
	return classrooms, pg, nil
}
