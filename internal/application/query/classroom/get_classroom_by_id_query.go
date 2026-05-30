package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
)

type GetClassroomByIdQuery struct {
	ClassroomID string
}

type GetClassroomByIdQueryHandler struct {
	classroomRepo classroom.IRepository
}

func NewGetClassroomByIdQueryHandler(classroomRepo classroom.IRepository) *GetClassroomByIdQueryHandler {
	return &GetClassroomByIdQueryHandler{classroomRepo: classroomRepo}
}

func (h *GetClassroomByIdQueryHandler) Handle(ctx context.Context, q GetClassroomByIdQuery) (*classroom.Classroom, error) {
	return h.classroomRepo.FindByClassroomId(ctx, q.ClassroomID)
}
