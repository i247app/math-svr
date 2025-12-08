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
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type GradeService struct {
	validator       validators.IGradeValidator
	repo            diRepo.IGradeRepository
	storageService  diSvc.IStorageService
	responseBuilder *utils.ResponseBuilder
	fileManager     *utils.FileManager
}

func NewGradeService(
	validator validators.IGradeValidator,
	repo diRepo.IGradeRepository,
	storageService diSvc.IStorageService,
) diSvc.IGradeService {
	responseBuilder := utils.NewResponseBuilder(storageService)
	fileManager := utils.NewFileManager(storageService)

	return &GradeService{
		validator:       validator,
		repo:            repo,
		storageService:  storageService,
		responseBuilder: responseBuilder,
		fileManager:     fileManager,
	}
}

func (s *GradeService) ListGrades(ctx context.Context, req *dto.ListGradeRequest) (status.Code, []*dto.GradeResponse, *pagination.Pagination, error) {
	params := diRepo.ListGradesParams{
		Search:    req.Search,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	grades, pagination, err := s.repo.List(ctx, params)
	if err != nil {
		return status.FAIL, nil, nil, err
	}

	if len(grades) == 0 {
		return status.SUCCESS, []*dto.GradeResponse{}, pagination, nil
	}

	// Build responses with presigned URLs using shared utility
	res := s.responseBuilder.BuildGradeResponses(ctx, grades)

	return status.SUCCESS, res, pagination, nil
}

func (s *GradeService) GetGradeByID(ctx context.Context, id string) (status.Code, *dto.GradeResponse, error) {
	logger := logger.GetLogger(ctx)

	grade, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.Errorf("failed to get grade by ID: %v", err)
		return status.FAIL, nil, err
	}
	if grade == nil {
		return status.NOT_FOUND, nil, err_svc.ErrGradeNotFound
	}

	// Build response with presigned URL using shared utility
	res := s.responseBuilder.BuildGradeResponse(ctx, grade)

	return status.SUCCESS, res, nil
}

func (s *GradeService) GetGradeByLabel(ctx context.Context, label string) (status.Code, *dto.GradeResponse, error) {
	grade, err := s.repo.FindByLabel(ctx, label)
	if err != nil {
		return status.FAIL, nil, err
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
		return status.FAIL, nil, err
	}
	if existingGrade != nil {
		return status.GRADE_ALREADY_EXISTS, nil, err_svc.ErrGradeAlreadyExists
	}

	// Handle icon upload using shared file manager
	iconKey, statusCode, err := s.fileManager.UploadFile(ctx, req.ImageFile, req.ImageFilename, req.ImageContentType, "grade")
	if err != nil {
		return statusCode, nil, err
	}

	gradeDomain := dto.BuildGradeDomainForCreate(req)

	// Set icon URL if uploaded
	if iconKey != nil {
		gradeDomain.SetImageKey(iconKey)
	}

	// Create grade without transaction (simple single table insert)
	_, err = s.repo.Create(ctx, nil, gradeDomain)
	if err != nil {
		return status.FAIL, nil, err
	}

	// Fetch the created grade
	grade, err := s.repo.FindByID(ctx, gradeDomain.ID())
	if err != nil {
		return status.FAIL, nil, err
	}

	// Build response using shared utility
	res := s.responseBuilder.BuildGradeResponse(ctx, grade)

	return status.SUCCESS, res, nil
}

func (s *GradeService) UpdateGrade(ctx context.Context, req *dto.UpdateGradeRequest) (status.Code, *dto.GradeResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateUpdateGradeRequest(req); err != nil {
		return statusCode, nil, err
	}

	existingGrade, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.FAIL, nil, err
	}
	if existingGrade == nil {
		return status.NOT_FOUND, nil, err_svc.ErrGradeNotFound
	}

	// If updating label, check for duplicates
	if req.Label != nil && *req.Label != existingGrade.Label() {
		duplicateGrade, err := s.repo.FindByLabel(ctx, *req.Label)
		if err != nil {
			return status.FAIL, nil, err
		}
		if duplicateGrade != nil {
			return status.GRADE_ALREADY_EXISTS, nil, err_svc.ErrGradeAlreadyExists
		}
	}
	// Handle avatar updates using shared file manager
	newImageKey, statusCode, err := s.fileManager.UpdateFile(
		ctx,
		existingGrade.ImageKey(),
		req.ImageFile,
		req.ImageFilename,
		req.ImageContentType,
		req.DeleteImage,
		"grade",
	)
	if err != nil {
		return statusCode, nil, err
	}

	gradeDomain := dto.BuildGradeDomainForUpdate(req)
	_, err = s.repo.Update(ctx, gradeDomain)
	if err != nil {
		return status.FAIL, nil, err
	}

	// Set avatar URL if uploaded
	if newImageKey != nil {
		gradeDomain.SetImageKey(newImageKey)
	}

	// Fetch the updated grade
	grade, err := s.repo.FindByID(ctx, gradeDomain.ID())
	if err != nil {
		return status.FAIL, nil, err
	}

	// Build response using shared utility
	res := s.responseBuilder.BuildGradeResponse(ctx, grade)

	return status.SUCCESS, res, nil
}

func (s *GradeService) DeleteGrade(ctx context.Context, id string) (status.Code, error) {
	grade, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return status.FAIL, err
	}
	if grade == nil {
		return status.NOT_FOUND, err_svc.ErrGradeNotFound
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return status.FAIL, err
	}

	return status.SUCCESS, nil
}

func (s *GradeService) ForceDeleteGrade(ctx context.Context, id string) (status.Code, error) {
	err := s.repo.ForceDelete(ctx, nil, id)
	if err != nil {
		return status.FAIL, err
	}

	return status.SUCCESS, nil
}
