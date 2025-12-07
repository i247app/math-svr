package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ISemesterService interface {
	ListSemesters(ctx context.Context, req *dto.ListSemesterRequest) (status.Code, []*dto.SemesterResponse, *pagination.Pagination, error)
	GetSemesterByID(ctx context.Context, id string) (status.Code, *dto.SemesterResponse, error)
	GetSemesterByName(ctx context.Context, name string) (status.Code, *dto.SemesterResponse, error)
	CreateSemester(ctx context.Context, req *dto.CreateSemesterRequest) (status.Code, *dto.SemesterResponse, error)
	UpdateSemester(ctx context.Context, req *dto.UpdateSemesterRequest) (status.Code, *dto.SemesterResponse, error)
	DeleteSemester(ctx context.Context, id string) (status.Code, error)
	ForceDeleteSemester(ctx context.Context, id string) (status.Code, error)
}
