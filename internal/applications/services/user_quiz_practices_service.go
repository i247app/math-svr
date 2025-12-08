package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"math-ai.com/math-ai/internal/applications/dto"
	helper "math-ai.com/math-ai/internal/applications/helpers/chatbox_helper"
	"math-ai.com/math-ai/internal/applications/validators"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
)

type userLatestQuizService struct {
	validator     validators.IUserLatestQuizValidator
	repo          diRepo.IUserQuizPracticesRepository
	profileSvc    diSvc.IProfileService
	chatboxSvc    diSvc.IChatBoxService
	jsonSanitizer *helper.JSONSanitizer
}

func NewUserLatestQuizService(
	validator validators.IUserLatestQuizValidator,
	repo diRepo.IUserQuizPracticesRepository,
	profileSvc diSvc.IProfileService,
	chatboxSvc diSvc.IChatBoxService,
) diSvc.IUserQuizPracticesService {
	return &userLatestQuizService{
		validator:     validator,
		repo:          repo,
		profileSvc:    profileSvc,
		chatboxSvc:    chatboxSvc,
		jsonSanitizer: helper.NewJSONSanitizer(),
	}
}

func (s *userLatestQuizService) GenerateQuiz(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error) {
	logger := logger.GetLogger(ctx)

	statusCode, user, err := s.profileSvc.FetchProfile(ctx, &dto.FetchProfileRequest{
		UID: req.UID,
	})
	if err != nil {
		logger.Errorf("Failed to fetch user profile: %v", err)
		return statusCode, nil, fmt.Errorf("failed to fetch user profile: %v", err)
	}

	// Build generate quiz from request
	conv := dto.BuildGenerateQuizFromRequest(ctx, req, user)

	// log prompt for debugging
	for _, msg := range conv.Messages() {
		if msg.Role() == "user" {
			logger.Infof("User prompt: %s", msg.Content())
		}
	}

	statusCode, res, err := s.chatboxSvc.Generate(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to generate quiz: %v", err)
		return statusCode, nil, fmt.Errorf("failed to generate quiz: %v", err)
	}

	// Save latest quiz for the user
	if res.Response != "" {
		_, uqp, err := s.GetUserQuizPraticeByUID(ctx, &dto.GetUserQuizPracticesByUIDRequest{
			UID: req.UID,
		})
		if err != nil {
			logger.Errorf("Failed to get latest quiz for user %s: %v", req.UID, err)
		}

		if res == nil {
			_, _, err := s.CreateUserQuizPratice(ctx, &dto.CreateUserQuizPracticesRequest{
				UID:       req.UID,
				Questions: res.Response,
				AIReview:  "",
			})
			if err != nil {
				logger.Errorf("Failed to create latest quiz for user %s: %v", req.UID, err)
			}
		} else {
			resetData := "?"
			_, _, err = s.UpdateUserQuizPratice(ctx, &dto.UpdateUserQuizPracticesRequest{
				// ID:        res.ID,
				UID:       uqp.UID,
				Questions: &res.Response,
				Answers:   &resetData,
				AIReview:  &resetData,
			})
			if err != nil {
				logger.Errorf("Failed to update latest quiz for user %s: %v", req.UID, err)
			}
		}
	}

	return status.SUCCESS, res, nil
}

func (s *userLatestQuizService) SubmitQuiz(ctx context.Context, req *dto.SubmitQuizRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAnswer], error) {
	logger := logger.GetLogger(ctx)

	jsonAnswers, err := json.Marshal(req.Answers)
	if err != nil {
		log.Fatalf("Error marshaling struct to JSON: %v", err)
	}

	answersStr := string(jsonAnswers)

	statusCode, ulq, err := s.UpdateUserQuizPratice(ctx, &dto.UpdateUserQuizPracticesRequest{
		// ID:      req.UserLatestQuizID,
		UID:     req.UID,
		Answers: &answersStr,
	})

	if err != nil {
		logger.Errorf("Failed to udpate latest quizzes: %v", err)
		return statusCode, nil, fmt.Errorf("failed to udpate latest quizzes: %v", err)
	}

	// Build generate quiz from request
	conv := dto.BuildSubmitQuizAnswerFromRequest(ctx, req, ulq)

	// log prompt for debugging
	for _, msg := range conv.Messages() {
		if msg.Role() == "user" {
			logger.Infof("User prompt: %s", msg.Content())
		}
	}

	statusCode, res, err := s.chatboxSvc.Submit(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to generate quiz: %v", err)
		return statusCode, nil, fmt.Errorf("failed to generate quiz: %v", err)
	}

	statusCode, _, err = s.UpdateUserQuizPratice(ctx, &dto.UpdateUserQuizPracticesRequest{
		// ID:       req.UserLatestQuizID,
		UID:      req.UID,
		AIReview: &res.Data.AIReview,
	})
	if err != nil {
		logger.Errorf("Failed to update user latest quiz with AI review: %v", err)
		return statusCode, nil, fmt.Errorf("failed to update user latest quiz with AI review: %v", err)
	}

	return status.SUCCESS, res, nil
}

func (s *userLatestQuizService) GenerateQuizPractice(ctx context.Context, req *dto.GenerateQuizPracticeRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error) {
	logger := logger.GetLogger(ctx)

	statusCode, ulq, err := s.GetUserQuizPraticeByUID(ctx, &dto.GetUserQuizPracticesByUIDRequest{
		UID: req.UID,
	})
	if err != nil {
		logger.Errorf("Failed to fetch user latest quiz: %v", err)
		return statusCode, nil, fmt.Errorf("failed to fetch user latest quiz: %v", err)
	}

	// Build generate practice quiz from request
	conv := dto.BuildGeneratePracticeQuizFromRequest(ctx, req, ulq)

	// log prompt for debugging
	for _, msg := range conv.Messages() {
		if msg.Role() == "user" {
			logger.Infof("User prompt: %s", msg.Content())
		}
	}

	statusCode, res, err := s.chatboxSvc.GeneratePractice(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to generate quiz: %v", err)
		return statusCode, nil, fmt.Errorf("failed to generate quiz: %v", err)
	}

	// Save latest quiz for the user
	if res.Response != "" {
		resetData := "?"
		_, _, err := s.UpdateUserQuizPratice(ctx, &dto.UpdateUserQuizPracticesRequest{
			ID:        ulq.ID,
			UID:       ulq.UID,
			Questions: &res.Response,
			Answers:   &resetData,
			AIReview:  &resetData,
		})
		if err != nil {
			logger.Errorf("Failed to update latest quiz for user %s: %v", req.UID, err)
		}
	}

	return status.SUCCESS, nil, nil
}

// GetQuiz retrieves a specific user latest quiz by ID.
func (s *userLatestQuizService) GetUserQuizPratice(ctx context.Context, req *dto.GetUserQuizPracticesRequest) (status.Code, *dto.UserQuizPracticesResponse, error) {
	quiz, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_FAILED, nil, err
	}

	if quiz == nil {
		return status.USER_LATEST_QUIZ_NOT_FOUND, nil, fmt.Errorf("user latest quiz not found")
	}

	response := dto.UserQuizPracticesResponseFromDomain(quiz)

	return status.OK, &response, nil
}

// GetQuizByUID retrieves the latest quiz for a specific user by UID.
func (s *userLatestQuizService) GetUserQuizPraticeByUID(ctx context.Context, req *dto.GetUserQuizPracticesByUIDRequest) (status.Code, *dto.UserQuizPracticesResponse, error) {
	quiz, err := s.repo.FindByUID(ctx, req.UID)
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_FAILED, nil, err
	}

	if quiz == nil {
		return status.USER_LATEST_QUIZ_NOT_FOUND, nil, fmt.Errorf("user latest quiz not found")
	}

	response := dto.UserQuizPracticesResponseFromDomain(quiz)

	return status.OK, &response, nil
}

// CreateQuiz creates a new user latest quiz.
func (s *userLatestQuizService) CreateUserQuizPratice(ctx context.Context, req *dto.CreateUserQuizPracticesRequest) (status.Code, *dto.UserQuizPracticesResponse, error) {
	// vaidate request
	if statusCode, err := s.validator.ValidateCreateUserLatestQuizRequest(req); err != nil {
		return statusCode, nil, err
	}

	quizDomain := dto.BuildUserQuizPracticesDomainForCreate(req)

	rowsAffected, err := s.repo.Create(ctx, nil, quizDomain)
	if err != nil {
		return status.USER_LATEST_QUIZ_CREATE_FAILED, nil, err
	}

	if rowsAffected == 0 {
		return status.USER_LATEST_QUIZ_CREATE_FAILED, nil, fmt.Errorf("failed to create user latest quiz")
	}

	response := dto.UserQuizPracticesResponseFromDomain(quizDomain)

	return status.CREATED, &response, nil
}

// UpdateQuiz updates an existing user latest quiz.
func (s *userLatestQuizService) UpdateUserQuizPratice(ctx context.Context, req *dto.UpdateUserQuizPracticesRequest) (status.Code, *dto.UserQuizPracticesResponse, error) {
	// vaidate request
	if statusCode, err := s.validator.ValidateUpdateUserLatestQuizRequest(req); err != nil {
		return statusCode, nil, err
	}

	existingQuiz, err := s.repo.FindByUID(ctx, req.UID)
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_FAILED, nil, err
	}

	if existingQuiz == nil {
		return status.USER_LATEST_QUIZ_NOT_FOUND, nil, fmt.Errorf("user latest quiz not found")
	}

	quizDomain := dto.BuildUserQuizPracticesDomainForUpdate(req)

	rowsAffected, err := s.repo.Update(ctx, quizDomain)
	if err != nil {
		return status.USER_LATEST_QUIZ_UPDATE_FAILED, nil, err
	}

	if rowsAffected == 0 {
		return status.USER_LATEST_QUIZ_UPDATE_FAILED, nil, fmt.Errorf("failed to update user latest quiz")
	}

	updatedQuiz, err := s.repo.FindByID(ctx, existingQuiz.ID())
	if err != nil {
		return status.USER_LATEST_QUIZ_GET_FAILED, nil, err
	}

	response := dto.UserQuizPracticesResponseFromDomain(updatedQuiz)

	return status.OK, &response, nil
}

// DeleteQuiz performs a soft delete on a user latest quiz.
func (s *userLatestQuizService) DeleteUserQuizPratice(ctx context.Context, req *dto.DeleteUserQuizPracticesRequest) (status.Code, error) {
	// vaidate request
	if statusCode, err := s.validator.ValidateDeleteUserLatestQuizRequest(req); err != nil {
		return statusCode, err
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
func (s *userLatestQuizService) ForceDeleteUserQuizPratice(ctx context.Context, req *dto.DeleteUserQuizPracticesRequest) (status.Code, error) {
	// vaidate request
	if statusCode, err := s.validator.ValidateDeleteUserLatestQuizRequest(req); err != nil {
		return statusCode, err
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
