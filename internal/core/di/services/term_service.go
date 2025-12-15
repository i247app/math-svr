package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ITermService interface {
	ListTerms(ctx context.Context, req *dto.ListTermRequest) (status.Code, []*dto.TermResponse, *pagination.Pagination, error)
	GetTermByID(ctx context.Context, id string) (status.Code, *dto.TermResponse, error)
	GetTermByName(ctx context.Context, name string) (status.Code, *dto.TermResponse, error)
	CreateTerm(ctx context.Context, req *dto.CreateTermRequest) (status.Code, *dto.TermResponse, error)
	UpdateTerm(ctx context.Context, req *dto.UpdateTermRequest) (status.Code, *dto.TermResponse, error)
	DeleteTerm(ctx context.Context, id string) (status.Code, error)
	ForceDeleteTerm(ctx context.Context, id string) (status.Code, error)
}
