package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ILevelService interface {
	ListLevels(ctx context.Context, req *dto.ListLevelRequest) (status.Code, []*dto.LevelResponse, *pagination.Pagination, error)
	GetLevelByID(ctx context.Context, id string) (status.Code, *dto.LevelResponse, error)
	GetLevelByLabel(ctx context.Context, label string) (status.Code, *dto.LevelResponse, error)
	CreateLevel(ctx context.Context, req *dto.CreateLevelRequest) (status.Code, *dto.LevelResponse, error)
	UpdateLevel(ctx context.Context, req *dto.UpdateLevelRequest) (status.Code, *dto.LevelResponse, error)
	DeleteLevel(ctx context.Context, id string) (status.Code, error)
	ForceDeleteLevel(ctx context.Context, id string) (status.Code, error)
}
