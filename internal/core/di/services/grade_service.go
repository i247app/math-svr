package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type IGradeService interface {
	ListGrades(ctx context.Context, req *dto.ListGradeRequest) (status.Code, []*dto.GradeResponse, *pagination.Pagination, error)
	GetGradeByID(ctx context.Context, id string) (status.Code, *dto.GradeResponse, error)
	GetGradeByLabel(ctx context.Context, label string) (status.Code, *dto.GradeResponse, error)
	CreateGrade(ctx context.Context, req *dto.CreateGradeRequest) (status.Code, *dto.GradeResponse, error)
	UpdateGrade(ctx context.Context, req *dto.UpdateGradeRequest) (status.Code, *dto.GradeResponse, error)
	DeleteGrade(ctx context.Context, id string) (status.Code, error)
	ForceDeleteGrade(ctx context.Context, id string) (status.Code, error)
}
