package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type ListProgramsQuery struct {
	Page  int64
	Limit int64
}

type ListProgramsQueryHandler struct {
	programRepo program.IRepository
}

func NewListProgramsQueryHandler(programRepo program.IRepository) *ListProgramsQueryHandler {
	return &ListProgramsQueryHandler{programRepo: programRepo}
}

func (h *ListProgramsQueryHandler) Handle(ctx context.Context, query *ListProgramsQuery) ([]*program.Program, *pagination.Pagination, error) {
	params := &program.ListProgramsParams{
		Page:  query.Page,
		Limit: query.Limit,
	}

	return h.programRepo.ListPrograms(ctx, params)
}
