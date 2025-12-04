package services

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	"math-ai.com/math-ai/internal/core/di/repositories"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
)

type ProfileService struct {
	validator      validators.IProfileValidator
	repo           repositories.IProfileRepository
	storageService di.IStorageService
}

func NewProfileService(
	validator validators.IProfileValidator,
	repo repositories.IProfileRepository,
	storageService di.IStorageService,
) di.IProfileService {
	return &ProfileService{
		validator:      validator,
		repo:           repo,
		storageService: storageService,
	}
}

func (s *ProfileService) FetchProfile(ctx context.Context, req *dto.FetchProfileRequest) (status.Code, *dto.ProfileResponse, error) {
	profile, err := s.repo.FindByUID(ctx, req.UID)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if profile == nil {
		return status.NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	res := dto.ProfileResponseFromDomain(profile)

	if profile.AvatarKey() != nil {
		statusCode, avatarURL, err := s.storageService.CreatePresignedUrl(ctx, &dto.CreatePresignedUrlRequest{
			Key:        *profile.AvatarKey(),
			Expiration: time.Hour,
		})
		if err != nil {
			return statusCode, nil, err
		}
		res.AvatarPreviewURL = &avatarURL
	}

	return status.SUCCESS, &res, nil
}

func (s *ProfileService) CreateProfile(ctx context.Context, req *dto.CreateProfileRequest) (status.Code, *dto.ProfileResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateCreateProfileRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Check if profile already exists for this user
	existingProfile, err := s.repo.FindByUID(ctx, req.UID)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if existingProfile != nil {
		return status.PROFILE_ALREADY_EXISTS, nil, err_svc.ErrProfileAlreadyExists
	}

	profileDomain := dto.BuildProfileDomainForCreate(req)

	// Create profile without transaction (simple single table insert)
	_, err = s.repo.Create(ctx, nil, profileDomain)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Fetch the created profile
	profile, err := s.repo.FindByID(ctx, profileDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res := dto.ProfileResponseFromDomain(profile)

	if profile.AvatarKey() != nil {
		statusCode, avatarURL, err := s.storageService.CreatePresignedUrl(ctx, &dto.CreatePresignedUrlRequest{
			Key:        *profile.AvatarKey(),
			Expiration: time.Hour,
		})
		if err != nil {
			return statusCode, nil, err
		}
		res.AvatarPreviewURL = &avatarURL
	}

	return status.SUCCESS, &res, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, req *dto.UpdateProfileRequest) (status.Code, *dto.ProfileResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateUpdateProfileRequest(req); err != nil {
		return statusCode, nil, err
	}

	profileDomain := dto.BuildProfileDomainForUpdate(req)
	_, err := s.repo.Update(ctx, profileDomain)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Fetch the updated profile
	profile, err := s.repo.FindByUID(ctx, profileDomain.UID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res := dto.ProfileResponseFromDomain(profile)

	return status.SUCCESS, &res, nil
}
