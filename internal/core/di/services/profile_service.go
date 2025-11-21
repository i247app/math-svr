package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IProfileService interface {
	FetchProfile(ctx context.Context, req *dto.FetchProfileRequest) (status.Code, *dto.ProfileResponse, error)
	CreateProfile(ctx context.Context, req *dto.CreateProfileRequest) (status.Code, *dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, req *dto.UpdateProfileRequest) (status.Code, *dto.ProfileResponse, error)
}
