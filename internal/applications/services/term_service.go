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

type TermService struct {
	validator       validators.ITermValidator
	repo            diRepo.ITermRepository
	storageService  diSvc.IStorageService
	responseBuilder *utils.ResponseBuilder
	fileManager     *utils.FileManager
}

func NewTermService(
	validator validators.ITermValidator,
	repo diRepo.ITermRepository,
	storageService diSvc.IStorageService,
) diSvc.ITermService {
	responseBuilder := utils.NewResponseBuilder(storageService)
	fileManager := utils.NewFileManager(storageService)

	return &TermService{
		validator:       validator,
		repo:            repo,
		storageService:  storageService,
		responseBuilder: responseBuilder,
		fileManager:     fileManager,
	}
}

func (s *TermService) ListTerms(ctx context.Context, req *dto.ListTermRequest) (status.Code, []*dto.TermResponse, *pagination.Pagination, error) {
	params := diRepo.ListTermsParams{
		Search:    req.Search,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	terms, pagination, err := s.repo.List(ctx, params)
	if err != nil {
		return status.FAIL, nil, nil, err
	}

	if len(terms) == 0 {
		return status.SUCCESS, []*dto.TermResponse{}, pagination, nil
	}

	// Build responses with presigned URLs using shared utility
	res := s.responseBuilder.BuildTermResponses(ctx, terms)

	return status.SUCCESS, res, pagination, nil
}

func (s *TermService) GetTermByID(ctx context.Context, id string) (status.Code, *dto.TermResponse, error) {
	logger := logger.GetLogger(ctx)

	term, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.Errorf("failed to get term by ID: ", err)
		return status.FAIL, nil, err
	}
	if term == nil {
		return status.NOT_FOUND, nil, err_svc.ErrTermNotFound
	}

	// Build response with presigned URL using shared utility
	res := s.responseBuilder.BuildTermResponse(ctx, term)

	return status.SUCCESS, res, nil
}

func (s *TermService) GetTermByName(ctx context.Context, name string) (status.Code, *dto.TermResponse, error) {
	term, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return status.FAIL, nil, err
	}
	if term == nil {
		return status.NOT_FOUND, nil, err_svc.ErrTermNotFound
	}

	res := dto.TermResponseFromDomain(term)

	return status.SUCCESS, &res, nil
}

func (s *TermService) CreateTerm(ctx context.Context, req *dto.CreateTermRequest) (status.Code, *dto.TermResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateCreateTermRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Check if term with same name already exists
	existingTerm, err := s.repo.FindByName(ctx, req.Name)
	if err != nil {
		return status.FAIL, nil, err
	}
	if existingTerm != nil {
		return status.TERM_ALREADY_EXISTS, nil, err_svc.ErrTermAlreadyExists
	}

	// Handle icon upload using shared file manager
	iconKey, statusCode, err := s.fileManager.UploadFile(ctx, req.ImageFile, req.ImageFilename, req.ImageContentType, "term")
	if err != nil {
		return statusCode, nil, err
	}

	termDomain := dto.BuildTermDomainForCreate(req)

	// Set icon URL if uploaded
	if iconKey != nil {
		termDomain.SetImageKey(iconKey)
	}

	// Create term without transaction (simple single table insert)
	_, err = s.repo.Create(ctx, nil, termDomain)
	if err != nil {
		return status.FAIL, nil, err
	}

	// Fetch the created term
	term, err := s.repo.FindByID(ctx, termDomain.ID())
	if err != nil {
		return status.FAIL, nil, err
	}

	// Build response using shared utility
	res := s.responseBuilder.BuildTermResponse(ctx, term)

	return status.SUCCESS, res, nil
}

func (s *TermService) UpdateTerm(ctx context.Context, req *dto.UpdateTermRequest) (status.Code, *dto.TermResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateUpdateTermRequest(req); err != nil {
		return statusCode, nil, err
	}

	existingTerm, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.FAIL, nil, err
	}
	if existingTerm == nil {
		return status.NOT_FOUND, nil, err_svc.ErrTermNotFound
	}

	// If updating name, check for duplicates
	if req.Name != nil && *req.Name != existingTerm.Name() {
		duplicateTerm, err := s.repo.FindByName(ctx, *req.Name)
		if err != nil {
			return status.FAIL, nil, err
		}
		if duplicateTerm != nil {
			return status.TERM_ALREADY_EXISTS, nil, err_svc.ErrTermAlreadyExists
		}
	}

	termDomain := dto.BuildTermDomainForUpdate(req)
	_, err = s.repo.Update(ctx, termDomain)
	if err != nil {
		return status.FAIL, nil, err
	}

	// Fetch the updated term
	term, err := s.repo.FindByID(ctx, termDomain.ID())
	if err != nil {
		return status.FAIL, nil, err
	}

	// Build response using shared utility
	res := s.responseBuilder.BuildTermResponse(ctx, term)

	return status.SUCCESS, res, nil
}

func (s *TermService) DeleteTerm(ctx context.Context, id string) (status.Code, error) {
	term, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return status.FAIL, err
	}
	if term == nil {
		return status.NOT_FOUND, err_svc.ErrTermNotFound
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return status.FAIL, err
	}

	return status.SUCCESS, nil
}

func (s *TermService) ForceDeleteTerm(ctx context.Context, id string) (status.Code, error) {
	err := s.repo.ForceDelete(ctx, nil, id)
	if err != nil {
		return status.FAIL, err
	}

	return status.SUCCESS, nil
}
