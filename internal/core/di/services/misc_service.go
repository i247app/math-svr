package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IMiscService interface {
	Ping() (status.Code, error)
	DetermineLocation(ctx context.Context, req *dto.LocationRequest) (status.Code, *dto.LocationResponse, error)
}
