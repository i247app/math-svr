package services

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type LevelService struct {
	validator validators.ILevelValidator
	repo      diRepo.ILevelRepository
}

func NewLevelService(
	validator validators.ILevelValidator,
	repo diRepo.ILevelRepository,
) diSvc.ILevelService {
	return &LevelService{
		validator: validator,
		repo:      repo,
	}
}

func (s *LevelService) ListLevels(ctx context.Context, req *dto.ListLevelRequest) (status.Code, []*dto.LevelResponse, *pagination.Pagination, error) {
	params := diRepo.ListLevelsParams{
		Search:    req.Search,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	levels, pagination, err := s.repo.List(ctx, params)
	if err != nil {
		return status.INTERNAL, nil, nil, err
	}

	if len(levels) == 0 {
		return status.SUCCESS, []*dto.LevelResponse{}, pagination, nil
	}

	res := make([]*dto.LevelResponse, len(levels))
	for i, level := range levels {
		levelRes := dto.LevelResponseFromDomain(level)
		res[i] = &levelRes
	}

	return status.SUCCESS, res, pagination, nil
}

func (s *LevelService) GetLevelByID(ctx context.Context, id string) (status.Code, *dto.LevelResponse, error) {
	level, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if level == nil {
		return status.NOT_FOUND, nil, err_svc.ErrLevelNotFound
	}

	res := dto.LevelResponseFromDomain(level)

	return status.SUCCESS, &res, nil
}

func (s *LevelService) GetLevelByLabel(ctx context.Context, label string) (status.Code, *dto.LevelResponse, error) {
	level, err := s.repo.FindByLabel(ctx, label)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if level == nil {
		return status.NOT_FOUND, nil, err_svc.ErrLevelNotFound
	}

	res := dto.LevelResponseFromDomain(level)

	return status.SUCCESS, &res, nil
}

func (s *LevelService) CreateLevel(ctx context.Context, req *dto.CreateLevelRequest) (status.Code, *dto.LevelResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateCreateLevelRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Check if level with same label already exists
	existingLevel, err := s.repo.FindByLabel(ctx, req.Label)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if existingLevel != nil {
		return status.LEVEL_ALREADY_EXISTS, nil, err_svc.ErrLevelAlreadyExists
	}

	levelDomain := dto.BuildLevelDomainForCreate(req)

	// Create level without transaction (simple single table insert)
	_, err = s.repo.Create(ctx, nil, levelDomain)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Fetch the created level
	level, err := s.repo.FindByID(ctx, levelDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res := dto.LevelResponseFromDomain(level)

	return status.SUCCESS, &res, nil
}

func (s *LevelService) UpdateLevel(ctx context.Context, req *dto.UpdateLevelRequest) (status.Code, *dto.LevelResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateUpdateLevelRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Check if level exists
	existingLevel, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if existingLevel == nil {
		return status.NOT_FOUND, nil, err_svc.ErrLevelNotFound
	}

	// If updating label, check for duplicates
	if req.Label != nil && *req.Label != existingLevel.Label() {
		duplicateLevel, err := s.repo.FindByLabel(ctx, *req.Label)
		if err != nil {
			return status.INTERNAL, nil, err
		}
		if duplicateLevel != nil {
			return status.LEVEL_ALREADY_EXISTS, nil, err_svc.ErrLevelAlreadyExists
		}
	}

	levelDomain := dto.BuildLevelDomainForUpdate(req)
	_, err = s.repo.Update(ctx, levelDomain)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	// Fetch the updated level
	level, err := s.repo.FindByID(ctx, levelDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res := dto.LevelResponseFromDomain(level)

	return status.SUCCESS, &res, nil
}

func (s *LevelService) DeleteLevel(ctx context.Context, id string) (status.Code, error) {
	// Check if level exists
	level, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return status.INTERNAL, err
	}
	if level == nil {
		return status.NOT_FOUND, err_svc.ErrLevelNotFound
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return status.INTERNAL, err
	}

	return status.SUCCESS, nil
}

func (s *LevelService) ForceDeleteLevel(ctx context.Context, id string) (status.Code, error) {
	err := s.repo.ForceDelete(ctx, nil, id)
	if err != nil {
		return status.INTERNAL, err
	}

	return status.SUCCESS, nil
}
