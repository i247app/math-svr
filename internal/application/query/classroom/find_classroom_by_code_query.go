package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
)

type FindClassroomByCodeQuery struct {
	ClassroomCode string
}

type FindClassroomByCodeQueryHandler struct {
	classroomRepo        classroom.IRepository
	classroomProgramRepo classroom.IClassroomProgramRepository
}

func NewFindClassroomByCodeQueryHandler(
	classroomRepo classroom.IRepository,
	classroomProgramRepo classroom.IClassroomProgramRepository,
) *FindClassroomByCodeQueryHandler {
	return &FindClassroomByCodeQueryHandler{
		classroomRepo:        classroomRepo,
		classroomProgramRepo: classroomProgramRepo,
	}
}

// Handle resolves the classroom by its join code. Nil result means
// not-found; on a hit the ProgramIds() slice is hydrated (possibly
// empty) so DomainToResponse renders a stable shape regardless of
// program membership.
func (h *FindClassroomByCodeQueryHandler) Handle(ctx context.Context, q FindClassroomByCodeQuery) (*classroom.Classroom, error) {
	found, err := h.classroomRepo.FindByClassroomCode(ctx, q.ClassroomCode)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, nil
	}
	programIDs, err := h.classroomProgramRepo.ListProgramIdsByClassroomId(ctx, found.ClassroomId())
	if err != nil {
		return nil, err
	}
	if programIDs == nil {
		programIDs = []int64{}
	}
	found.SetProgramIds(programIDs)
	return found, nil
}
