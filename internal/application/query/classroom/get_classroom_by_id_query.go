package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
)

type GetClassroomByIdQuery struct {
	ClassroomID int64
}

type GetClassroomByIdQueryHandler struct {
	classroomRepo        classroom.IRepository
	classroomProgramRepo classroom.IClassroomProgramRepository
}

func NewGetClassroomByIdQueryHandler(
	classroomRepo classroom.IRepository,
	classroomProgramRepo classroom.IClassroomProgramRepository,
) *GetClassroomByIdQueryHandler {
	return &GetClassroomByIdQueryHandler{
		classroomRepo:        classroomRepo,
		classroomProgramRepo: classroomProgramRepo,
	}
}

// Handle returns the classroom with its program link set already
// hydrated into the in-memory Classroom struct. A nil result means
// not-found; on a hit the ProgramIds() slice is non-nil even when the
// classroom has zero programs, so DomainToResponse can distinguish
// "no programs" (`[]`) from "not hydrated" (`null`).
func (h *GetClassroomByIdQueryHandler) Handle(ctx context.Context, q GetClassroomByIdQuery) (*classroom.Classroom, error) {
	found, err := h.classroomRepo.FindByClassroomId(ctx, q.ClassroomID)
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
