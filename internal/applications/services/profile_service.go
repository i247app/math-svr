package services

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/utils"
	"math-ai.com/math-ai/internal/applications/validators"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/telemetry/metrics"
)

type ProfileService struct {
	validator       validators.IProfileValidator
	repo            diRepo.IProfileRepository
	storageService  diSvc.IStorageService
	responseBuilder *utils.ResponseBuilder
	fileManager     *utils.FileManager
}

func NewProfileService(
	validator validators.IProfileValidator,
	repo diRepo.IProfileRepository,
	storageService diSvc.IStorageService,
) diSvc.IProfileService {
	responseBuilder := utils.NewResponseBuilder(storageService)
	fileManager := utils.NewFileManager(storageService)

	return &ProfileService{
		validator:       validator,
		repo:            repo,
		storageService:  storageService,
		responseBuilder: responseBuilder,
		fileManager:     fileManager,
	}
}

func (s *ProfileService) FetchProfile(ctx context.Context, req *dto.FetchProfileRequest) (status.Code, *dto.ProfileResponse, error) {
	profile, err := s.repo.FindByUID(ctx, req.UID)
	if err != nil {
		return status.FAIL, nil, err
	}
	if profile == nil {
		return status.NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	// Build response with presigned URL using shared utility
	res := s.responseBuilder.BuildProfileResponse(ctx, profile)

	return status.SUCCESS, res, nil
}

func (s *ProfileService) CreateProfile(ctx context.Context, req *dto.CreateProfileRequest) (status.Code, *dto.ProfileResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateCreateProfileRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Check if profile already exists for this user
	existingProfile, err := s.repo.FindByUID(ctx, req.UID)
	if err != nil {
		return status.FAIL, nil, err
	}
	if existingProfile != nil {
		return status.PROFILE_ALREADY_EXISTS, nil, err_svc.ErrProfileAlreadyExists
	}

	profileDomain := dto.BuildProfileDomainForCreate(req)

	// Create profile without transaction (simple single table insert)
	_, err = s.repo.Create(ctx, nil, profileDomain)
	if err != nil {
		return status.FAIL, nil, err
	}

	// Fetch the created profile
	profile, err := s.repo.FindByID(ctx, profileDomain.ID())
	if err != nil {
		return status.FAIL, nil, err
	}

	// Build response with presigned URL using shared utility
	res := s.responseBuilder.BuildProfileResponse(ctx, profile)

	// Record business metric for profile creation
	gradeLevel := profile.Grade()
	metrics.RecordProfileCreation(gradeLevel)

	return status.SUCCESS, res, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, req *dto.UpdateProfileRequest) (status.Code, *dto.ProfileResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateUpdateProfileRequest(req); err != nil {
		return statusCode, nil, err
	}

	profileDomain := dto.BuildProfileDomainForUpdate(req)
	_, err := s.repo.Update(ctx, nil, profileDomain)
	if err != nil {
		return status.FAIL, nil, err
	}

	// Fetch the updated profile
	profile, err := s.repo.FindByUID(ctx, profileDomain.UID())
	if err != nil {
		return status.FAIL, nil, err
	}

	res := dto.ProfileResponseFromDomain(profile)

	// Record business metric for profile update
	updateType := "general"
	if req.GradeID != nil {
		updateType = "grade"
	} else if req.SemesterID != nil {
		updateType = "semester"
	}
	metrics.RecordProfileUpdate(updateType)

	return status.SUCCESS, &res, nil
}
