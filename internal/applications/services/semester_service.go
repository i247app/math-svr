package services

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/utils"
	"math-ai.com/math-ai/internal/applications/validators"
	di "math-ai.com/math-ai/internal/core/di/repositories"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type SemesterService struct {
	validator       validators.ISemesterValidator
	repo            diRepo.ISemesterRepository
	storageService  diSvc.IStorageService
	responseBuilder *utils.ResponseBuilder
	fileManager     *utils.FileManager
}

func NewSemesterService(
	validator validators.ISemesterValidator,
	repo diRepo.ISemesterRepository,
	storageService diSvc.IStorageService,
) diSvc.ISemesterService {
	responseBuilder := utils.NewResponseBuilder(storageService)
	fileManager := utils.NewFileManager(storageService)

	return &SemesterService{
		validator:       validator,
		repo:            repo,
		storageService:  storageService,
		responseBuilder: responseBuilder,
		fileManager:     fileManager,
	}
}

func (s *SemesterService) ListSemesters(ctx context.Context, req *dto.ListSemesterRequest) (status.Code, []*dto.SemesterResponse, *pagination.Pagination, error) {
	params := di.ListSemestersParams{
		Search:    req.Search,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	semesters, pagination, err := s.repo.List(ctx, params)
	if err != nil {
		return status.INTERNAL, nil, nil, err
	}

	if len(semesters) == 0 {
		return status.SUCCESS, []*dto.SemesterResponse{}, pagination, nil
	}

	// Build responses with presigned URLs using shared utility
	res := s.responseBuilder.BuildSemesterResponses(ctx, semesters)

	return status.SUCCESS, res, pagination, nil
}

func (s *SemesterService) GetSemesterByID(ctx context.Context, id string) (status.Code, *dto.SemesterResponse, error) {
	logger := logger.GetLogger(ctx)

	semester, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.Errorf("failed to get semester by ID: ", err)
		return status.INTERNAL, nil, err
	}
	if semester == nil {
		return status.NOT_FOUND, nil, err_svc.ErrSemesterNotFound
	}

	// Build response with presigned URL using shared utility
	res := s.responseBuilder.BuildSemesterResponse(ctx, semester)

	return status.SUCCESS, res, nil
}

func (s *SemesterService) GetSemesterByName(ctx context.Context, name string) (status.Code, *dto.SemesterResponse, error) {
	semester, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if semester == nil {
		return status.NOT_FOUND, nil, err_svc.ErrSemesterNotFound
	}

	res := dto.SemesterResponseFromDomain(semester)

	return status.SUCCESS, &res, nil
}

func (s *SemesterService) CreateSemester(ctx context.Context, req *dto.CreateSemesterRequest) (status.Code, *dto.SemesterResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateCreateSemesterRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Check if semester with same name already exists
	existingSemester, err := s.repo.FindByName(ctx, req.Name)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if existingSemester != nil {
		return status.SEMESTER_ALREADY_EXISTS, nil, err_svc.ErrSemesterAlreadyExists
	}

	// Handle icon upload using shared file manager
	iconKey, statusCode, err := s.fileManager.UploadFile(ctx, req.IconFile, req.IconFilename, req.IconContentType, "semester")
	if err != nil {
		return statusCode, nil, err
	}

	semesterDomain := dto.BuildSemesterDomainForCreate(req)

	// Set icon URL if uploaded
	if iconKey != nil {
		semesterDomain.SetImageKey(iconKey)
	}

	// Create semester without transaction (simple single table insert)
	_, err = s.repo.Create(ctx, nil, semesterDomain)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Fetch the created semester
	semester, err := s.repo.FindByID(ctx, semesterDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Build response using shared utility
	res := s.responseBuilder.BuildSemesterResponse(ctx, semester)

	return status.SUCCESS, res, nil
}

func (s *SemesterService) UpdateSemester(ctx context.Context, req *dto.UpdateSemesterRequest) (status.Code, *dto.SemesterResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateUpdateSemesterRequest(req); err != nil {
		return statusCode, nil, err
	}

	existingSemester, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if existingSemester == nil {
		return status.NOT_FOUND, nil, err_svc.ErrSemesterNotFound
	}

	// If updating name, check for duplicates
	if req.Name != nil && *req.Name != existingSemester.Name() {
		duplicateSemester, err := s.repo.FindByName(ctx, *req.Name)
		if err != nil {
			return status.INTERNAL, nil, err
		}
		if duplicateSemester != nil {
			return status.SEMESTER_ALREADY_EXISTS, nil, err_svc.ErrSemesterAlreadyExists
		}
	}

	semesterDomain := dto.BuildSemesterDomainForUpdate(req)
	_, err = s.repo.Update(ctx, semesterDomain)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Fetch the updated semester
	semester, err := s.repo.FindByID(ctx, semesterDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Build response using shared utility
	res := s.responseBuilder.BuildSemesterResponse(ctx, semester)

	return status.SUCCESS, res, nil
}

func (s *SemesterService) DeleteSemester(ctx context.Context, id string) (status.Code, error) {
	semester, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return status.INTERNAL, err
	}
	if semester == nil {
		return status.NOT_FOUND, err_svc.ErrSemesterNotFound
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return status.INTERNAL, err
	}

	return status.SUCCESS, nil
}

func (s *SemesterService) ForceDeleteSemester(ctx context.Context, id string) (status.Code, error) {
	err := s.repo.ForceDelete(ctx, nil, id)
	if err != nil {
		return status.INTERNAL, err
	}

	return status.SUCCESS, nil
}
