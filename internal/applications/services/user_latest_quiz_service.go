package services

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/core/di/repositories"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type userLatestQuizService struct {
	repo repositories.IUserLatestQuizRepository
}

func NewUserLatestQuizService(repo repositories.IUserLatestQuizRepository) di.IUserLatestQuizService {
	return &userLatestQuizService{
		repo: repo,
	}
}

// ListQuizzes retrieves a paginated list of user latest quizzes.
func (s *userLatestQuizService) ListQuizzes(ctx context.Context, req *dto.ListUserLatestQuizzesRequest) (status.Code, []*dto.UserLatestQuizResponse, *pagination.Pagination, error) {
	limit := req.Limit
	offset := req.Offset

	if limit <= 0 {
		limit = 10
	}

	quizzes, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_LIST_FAILED, nil, nil, err
	}

	if len(quizzes) == 0 {
		return status.SUCCESS, []*dto.UserLatestQuizResponse{}, &pagination.Pagination{}, nil
	}

	res := make([]*dto.UserLatestQuizResponse, len(quizzes))
	for i, quiz := range quizzes {
		quizRes := dto.UserLatestQuizResponseFromDomain(quiz)
		res[i] = &quizRes
	}

	page := int64(offset/limit) + 1
	if offset == 0 && limit > 0 {
		page = 1
	}

	pag := &pagination.Pagination{
		Page:       page,
		Size:       int64(limit),
		Skip:       int64(offset),
		TotalCount: int64(len(quizzes)),
	}

	return status.SUCCESS, res, pag, nil
}

// GetQuiz retrieves a specific user latest quiz by ID.
func (s *userLatestQuizService) GetQuiz(ctx context.Context, req *dto.GetUserLatestQuizRequest) (status.Code, *dto.UserLatestQuizResponse, error) {
	if req.ID == "" {
		return status.BAD_REQUEST, nil, fmt.Errorf("id is required")
	}

	quiz, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_FAILED, nil, err
	}

	if quiz == nil {
		return status.USER_LATEST_QUIZ_NOT_FOUND, nil, fmt.Errorf("user latest quiz not found")
	}

	response := dto.UserLatestQuizResponseFromDomain(quiz)

	return status.OK, &response, nil
}

// GetQuizByUID retrieves the latest quiz for a specific user by UID.
func (s *userLatestQuizService) GetQuizByUID(ctx context.Context, req *dto.GetUserLatestQuizByUIDRequest) (status.Code, *dto.UserLatestQuizResponse, error) {
	if req.UID == "" {
		return status.BAD_REQUEST, nil, fmt.Errorf("uid is required")
	}

	quiz, err := s.repo.FindByUID(ctx, req.UID)
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_FAILED, nil, err
	}

	if quiz == nil {
		return status.USER_LATEST_QUIZ_NOT_FOUND, nil, fmt.Errorf("user latest quiz not found")
	}

	response := dto.UserLatestQuizResponseFromDomain(quiz)

	return status.OK, &response, nil
}

// CreateQuiz creates a new user latest quiz.
func (s *userLatestQuizService) CreateQuiz(ctx context.Context, req *dto.CreateUserLatestQuizRequest) (status.Code, *dto.UserLatestQuizResponse, error) {
	if req.UID == "" {
		return status.BAD_REQUEST, nil, fmt.Errorf("uid is required")
	}

	if req.Questions == "" {
		return status.BAD_REQUEST, nil, fmt.Errorf("questions is required")
	}

	if req.Answers == "" {
		return status.BAD_REQUEST, nil, fmt.Errorf("answers is required")
	}

	if req.AIReview == "" {
		return status.BAD_REQUEST, nil, fmt.Errorf("ai_review is required")
	}

	quizDomain := dto.BuildUserLatestQuizDomainForCreate(req)

	rowsAffected, err := s.repo.Create(ctx, nil, quizDomain)
	if err != nil {
		return status.USER_LATEST_QUIZ_CREATE_FAILED, nil, err
	}

	if rowsAffected == 0 {
		return status.USER_LATEST_QUIZ_CREATE_FAILED, nil, fmt.Errorf("failed to create user latest quiz")
	}

	response := dto.UserLatestQuizResponseFromDomain(quizDomain)

	return status.CREATED, &response, nil
}

// UpdateQuiz updates an existing user latest quiz.
func (s *userLatestQuizService) UpdateQuiz(ctx context.Context, req *dto.UpdateUserLatestQuizRequest) (status.Code, *dto.UserLatestQuizResponse, error) {
	if req.ID == "" {
		return status.BAD_REQUEST, nil, fmt.Errorf("id is required")
	}

	existingQuiz, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_FAILED, nil, err
	}

	if existingQuiz == nil {
		return status.USER_LATEST_QUIZ_NOT_FOUND, nil, fmt.Errorf("user latest quiz not found")
	}

	quizDomain := dto.BuildUserLatestQuizDomainForUpdate(req)

	rowsAffected, err := s.repo.Update(ctx, quizDomain)
	if err != nil {
		return status.USER_LATEST_QUIZ_UPDATE_FAILED, nil, err
	}

	if rowsAffected == 0 {
		return status.USER_LATEST_QUIZ_UPDATE_FAILED, nil, fmt.Errorf("failed to update user latest quiz")
	}

	updatedQuiz, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_FAILED, nil, err
	}

	response := dto.UserLatestQuizResponseFromDomain(updatedQuiz)

	return status.OK, &response, nil
}

// DeleteQuiz performs a soft delete on a user latest quiz.
func (s *userLatestQuizService) DeleteQuiz(ctx context.Context, req *dto.DeleteUserLatestQuizRequest) (status.Code, error) {
	if req.ID == "" {
		return status.BAD_REQUEST, fmt.Errorf("id is required")
	}

	existingQuiz, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_FAILED, err
	}

	if existingQuiz == nil {
		return status.USER_LATEST_QUIZ_NOT_FOUND, fmt.Errorf("user latest quiz not found")
	}

	rowsAffected, err := s.repo.Delete(ctx, req.ID)
	if err != nil {
		return status.USER_LATEST_QUIZ_DELETE_FAILED, err
	}

	if rowsAffected == 0 {
		return status.USER_LATEST_QUIZ_DELETE_FAILED, fmt.Errorf("failed to delete user latest quiz")
	}

	return status.OK, nil
}

// ForceDeleteQuiz permanently removes a user latest quiz.
func (s *userLatestQuizService) ForceDeleteQuiz(ctx context.Context, req *dto.ForceDeleteUserLatestQuizRequest) (status.Code, error) {
	if req.ID == "" {
		return status.BAD_REQUEST, fmt.Errorf("id is required")
	}

	rowsAffected, err := s.repo.ForceDelete(ctx, req.ID)
	if err != nil {
		return status.USER_LATEST_QUIZ_FORCE_DELETE_FAILED, err
	}

	if rowsAffected == 0 {
		return status.USER_LATEST_QUIZ_FORCE_DELETE_FAILED, fmt.Errorf("failed to force delete user latest quiz")
	}

	return status.OK, nil
}
