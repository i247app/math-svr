package services

import (
	"context"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	"math-ai.com/math-ai/internal/core/di/repositories"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type GradeService struct {
	validator      validators.IGradeValidator
	repo           repositories.IGradeRepository
	storageService di.IStorageService
}

func NewGradeService(
	validator validators.IGradeValidator,
	repo repositories.IGradeRepository,
	storageService di.IStorageService,
) di.IGradeService {
	return &GradeService{
		validator:      validator,
		repo:           repo,
		storageService: storageService,
	}
}

func (s *GradeService) ListGrades(ctx context.Context, req *dto.ListGradeRequest) (status.Code, []*dto.GradeResponse, *pagination.Pagination, error) {
	params := repositories.ListGradesParams{
		Search:    req.Search,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	grades, pagination, err := s.repo.List(ctx, params)
	if err != nil {
		return status.INTERNAL, nil, nil, err
	}

	if len(grades) == 0 {
		return status.SUCCESS, []*dto.GradeResponse{}, pagination, nil
	}

	res := make([]*dto.GradeResponse, len(grades))
	for i, grade := range grades {
		gradeRes := dto.GradeResponseFromDomain(grade)
		if grade.IconURL() != nil {
			signedURL, err := s.storageService.CreatePresignedUrl(ctx, *grade.IconURL(), time.Hour)
			if err != nil {
				return status.INTERNAL, nil, nil, fmt.Errorf("failed to create presigned URL: %w", err)
			}
			if signedURL != "" {
				gradeRes.IconURL = &signedURL
			}
		}
		res[i] = &gradeRes
	}

	return status.SUCCESS, res, pagination, nil
}

func (s *GradeService) GetGradeByID(ctx context.Context, id string) (status.Code, *dto.GradeResponse, error) {
	grade, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if grade == nil {
		return status.NOT_FOUND, nil, err_svc.ErrGradeNotFound
	}

	res := dto.GradeResponseFromDomain(grade)

	if grade.IconURL() != nil {
		signedURL, err := s.storageService.CreatePresignedUrl(ctx, *grade.IconURL(), time.Hour)
		if err != nil {
			return status.INTERNAL, nil, fmt.Errorf("failed to create presigned URL: %w", err)
		}
		if signedURL != "" {
			res.IconURL = &signedURL
		}
	}

	return status.SUCCESS, &res, nil
}

func (s *GradeService) GetGradeByLabel(ctx context.Context, label string) (status.Code, *dto.GradeResponse, error) {
	grade, err := s.repo.FindByLabel(ctx, label)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if grade == nil {
		return status.NOT_FOUND, nil, err_svc.ErrGradeNotFound
	}

	res := dto.GradeResponseFromDomain(grade)

	return status.SUCCESS, &res, nil
}

func (s *GradeService) CreateGrade(ctx context.Context, req *dto.CreateGradeRequest) (status.Code, *dto.GradeResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateCreateGradeRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Check if grade with same label already exists
	existingGrade, err := s.repo.FindByLabel(ctx, req.Label)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if existingGrade != nil {
		return status.GRADE_ALREADY_EXISTS, nil, err_svc.ErrGradeAlreadyExists
	}

	// Handle avatar upload before creating user
	var iconKey *string
	if req.IconFile != nil {
		res, err := s.storageService.HandleUpload(ctx, req.IconFile, req.IconFilename, req.IconContentType, "grade")
		if err != nil {
			return status.INTERNAL, nil, fmt.Errorf("failed to upload avatar: %w", err)
		}
		iconKey = &res.Key
	}

	gradeDomain := dto.BuildGradeDomainForCreate(req)

	// Set avatar URL if uploaded
	if iconKey != nil {
		gradeDomain.SetIconURL(iconKey)
	}

	// Create grade without transaction (simple single table insert)
	_, err = s.repo.Create(ctx, nil, gradeDomain)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Fetch the created grade
	grade, err := s.repo.FindByID(ctx, gradeDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res := dto.GradeResponseFromDomain(grade)

	return status.SUCCESS, &res, nil
}

func (s *GradeService) UpdateGrade(ctx context.Context, req *dto.UpdateGradeRequest) (status.Code, *dto.GradeResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateUpdateGradeRequest(req); err != nil {
		return statusCode, nil, err
	}

	existingGrade, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if existingGrade == nil {
		return status.NOT_FOUND, nil, err_svc.ErrGradeNotFound
	}

	// If updating label, check for duplicates
	if req.Label != nil && *req.Label != existingGrade.Label() {
		duplicateGrade, err := s.repo.FindByLabel(ctx, *req.Label)
		if err != nil {
			return status.INTERNAL, nil, err
		}
		if duplicateGrade != nil {
			return status.GRADE_ALREADY_EXISTS, nil, err_svc.ErrGradeAlreadyExists
		}
	}

	gradeDomain := dto.BuildGradeDomainForUpdate(req)
	_, err = s.repo.Update(ctx, gradeDomain)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Fetch the updated grade
	grade, err := s.repo.FindByID(ctx, gradeDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res := dto.GradeResponseFromDomain(grade)

	return status.SUCCESS, &res, nil
}

func (s *GradeService) DeleteGrade(ctx context.Context, id string) (status.Code, error) {
	grade, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return status.INTERNAL, err
	}
	if grade == nil {
		return status.NOT_FOUND, err_svc.ErrGradeNotFound
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return status.INTERNAL, err
	}

	return status.SUCCESS, nil
}

func (s *GradeService) ForceDeleteGrade(ctx context.Context, id string) (status.Code, error) {
	err := s.repo.ForceDelete(ctx, nil, id)
	if err != nil {
		return status.INTERNAL, err
	}

	return status.SUCCESS, nil
}
