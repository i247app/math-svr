package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/program"
)

// GetProgramByIdQuery returns the program row.
type GetProgramByIdQuery struct {
	ProgramID int64
}

type GetProgramByIdQueryHandler struct {
	programRepo program.IRepository
}

func NewGetProgramByIdQueryHandler(programRepo program.IRepository) *GetProgramByIdQueryHandler {
	return &GetProgramByIdQueryHandler{
		programRepo: programRepo,
	}
}

func (h *GetProgramByIdQueryHandler) Handle(ctx context.Context, q GetProgramByIdQuery) (*program.Program, error) {
	return h.programRepo.FindByProgramId(ctx, q.ProgramID)
}
