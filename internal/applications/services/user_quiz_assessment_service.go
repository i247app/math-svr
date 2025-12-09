package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"math-ai.com/math-ai/internal/applications/dto"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
)

type userQuizAssessmentService struct {
	repo       diRepo.IUserQuizAssessmentRepository
	profileSvc diSvc.IProfileService
	chatboxSvc diSvc.IChatBoxService
}

func NewUserQuizAssessmentService(
	repo diRepo.IUserQuizAssessmentRepository,
	profileSvc diSvc.IProfileService,
	chatboxSvc diSvc.IChatBoxService,
) diSvc.IUserQuizAssessmentService {
	return &userQuizAssessmentService{
		repo:       repo,
		profileSvc: profileSvc,
		chatboxSvc: chatboxSvc,
	}
}

func (s *userQuizAssessmentService) GenerateQuizAssessment(ctx context.Context, req *dto.GenerateQuizAssessmentRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error) {
	logger := logger.GetLogger(ctx)

	statusCode, user, err := s.profileSvc.FetchProfile(ctx, &dto.FetchProfileRequest{
		UID: req.UID,
	})
	if err != nil {
		logger.Errorf("Failed to fetch user profile: %v", err)
		return statusCode, nil, fmt.Errorf("failed to fetch user profile: %v", err)
	}

	// Build generate quiz from request
	conv := dto.BuildChatDomainGenerateQuizAssessment(ctx, &dto.GenerateQuizRequest{
		UID:                  req.UID,
		Grade:                req.Grade,
		ChatBoxRequestCommon: req.ChatBoxRequestCommon,
	}, user)

	// log prompt for debugging
	for _, msg := range conv.Messages() {
		if msg.Role() == "user" {
			logger.Infof("User prompt: %s", msg.Content())
		}
	}

	statusCode, res, err := s.chatboxSvc.Generate(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to generate quiz assessment: %v", err)
		return statusCode, nil, fmt.Errorf("failed to generate quiz assessment: %v", err)
	}

	data, err := json.Marshal(res.Data)
	if err != nil {
		logger.Errorf("Failed to marshal generated questions: %v", err)
		return status.FAIL, nil, fmt.Errorf("failed to marshal generated questions: %v", err)
	}

	statusCode, uqa, err := s.CreateUserQuizAssessment(ctx, &dto.CreateUserQuizAssessmentRequest{
		UID:       req.UID,
		Questions: string(data),
		Answers:   "",
		AIReview:  "",
	})
	if err != nil {
		logger.Errorf("Failed to create latest quiz for user %s: %v", req.UID, err)
		return statusCode, nil, fmt.Errorf("failed to create latest quiz for user %s: %v", req.UID, err)
	}

	res.UserQuizAssessmentID = uqa.ID

	return status.SUCCESS, res, nil
}

func (s *userQuizAssessmentService) SubmitQuizAssessment(ctx context.Context, req *dto.SubmitQuizAssessmentRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAssessmentAnswer], error) {
	logger := logger.GetLogger(ctx)

	jsonAnswers, err := json.Marshal(req.Answers)
	if err != nil {
		log.Fatalf("Error marshaling struct to JSON: %v", err)
	}

	answersStr := string(jsonAnswers)

	statusCode, uqa, err := s.UpdateUserQuizAssessment(ctx, &dto.UpdateUserQuizAssessmentRequest{
		ID:      req.UserQuizAssessmentID,
		Answers: &answersStr,
	})

	if err != nil {
		logger.Errorf("Failed to udpate latest quizzes: %v", err)
		return statusCode, nil, fmt.Errorf("failed to udpate latest quizzes: %v", err)
	}

	// Build submit quiz answer with assessment
	conv := dto.BuildChatDomainSubmitQuizAssessment(ctx, req, uqa)

	// log prompt for debugging
	for _, msg := range conv.Messages() {
		if msg.Role() == "user" {
			logger.Infof("User prompt: %s", msg.Content())
		}
	}

	statusCode, res, err := s.chatboxSvc.SubmitAssessment(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to submit quiz assessment: %v", err)
		return statusCode, nil, fmt.Errorf("failed to submit quiz assessment: %v", err)
	}

	// Save the assessment with AI-detected grade
	if res != nil && res.Data.AIReview != "" {
		statusCode, _, err := s.UpdateUserQuizAssessment(ctx, &dto.UpdateUserQuizAssessmentRequest{
			ID:            uqa.ID,
			AIReview:      &res.Data.AIReview,
			AIDetectGrade: &res.Data.AIDetectGrade,
		})

		if err != nil {
			logger.Errorf("Failed to udpate latest quizzes: %v", err)
			return statusCode, nil, fmt.Errorf("failed to udpate latest quizzes: %v", err)
		}
	}

	return status.SUCCESS, res, nil
}

func (s *userQuizAssessmentService) ReinforceQuizAssessment(ctx context.Context, req *dto.ReinforceQuizAssessmentRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error) {
	logger := logger.GetLogger(ctx)

	// Get the original assessment
	assessment, err := s.repo.FindByID(ctx, req.UserQuizAssessmentID)
	if err != nil {
		logger.Errorf("Failed to fetch quiz assessment: %v", err)
		return status.FAIL, nil, fmt.Errorf("failed to fetch quiz assessment: %v", err)
	}

	if assessment == nil {
		logger.Errorf("Quiz assessment not found: %s", req.UserQuizAssessmentID)
		return status.NOT_FOUND, nil, fmt.Errorf("quiz assessment not found")
	}

	assessmentResp := dto.UserQuizAssessmentResponseFromDomain(assessment)

	// Build generate practice quiz from request
	conv := dto.BuildChatDomainReinforceQuizAssessment(ctx, req, &assessmentResp)

	// log prompt for debugging
	for _, msg := range conv.Messages() {
		if msg.Role() == "user" {
			logger.Infof("User prompt: %s", msg.Content())
		}
	}

	statusCode, res, err := s.chatboxSvc.Reinforce(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to generate reinforcement quiz: %v", err)
		return statusCode, nil, fmt.Errorf("failed to generate reinforcement quiz: %v", err)
	}

	return status.SUCCESS, res, nil
}

func (s *userQuizAssessmentService) SubmitReinforceQuizAssessment(ctx context.Context, req *dto.SubmitReinforceQuizAssessmentRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAssessmentAnswer], error) {
	logger := logger.GetLogger(ctx)

	// Get the original assessment
	statusCode, assessment, err := s.GetUserQuizAssessmentByID(ctx, req.UserQuizAssessmentID)
	if err != nil {
		logger.Errorf("Failed to fetch quiz assessment: %v", err)
		return statusCode, nil, fmt.Errorf("failed to fetch quiz assessment: %v", err)
	}

	if assessment == nil {
		logger.Errorf("Quiz assessment not found: %s", req.UserQuizAssessmentID)
		return status.NOT_FOUND, nil, fmt.Errorf("quiz assessment not found")
	}

	// Get user profile
	statusCode, _, err = s.profileSvc.FetchProfile(ctx, &dto.FetchProfileRequest{
		UID: req.UID,
	})
	if err != nil {
		logger.Errorf("Failed to fetch user profile: %v", err)
		return statusCode, nil, fmt.Errorf("failed to fetch user profile: %v", err)
	}

	jsonAnswers, err := json.Marshal(req.Answers)
	if err != nil {
		log.Fatalf("Error marshaling struct to JSON: %v", err)
	}

	answersStr := string(jsonAnswers)

	// Build submit quiz answer with assessment
	conv := dto.BuildChatDomainSubmitReinforceQuizAssessment(ctx, &dto.ReinforceQuizAssessmentRequest{
		UID:                  req.UID,
		ChatBoxRequestCommon: req.ChatBoxRequestCommon,
	}, assessment)

	// log prompt for debugging
	for _, msg := range conv.Messages() {
		if msg.Role() == "user" {
			logger.Infof("User prompt: %s", msg.Content())
		}
	}

	statusCode, res, err := s.chatboxSvc.SubmitAssessment(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to submit reinforcement quiz: %v", err)
		return statusCode, nil, fmt.Errorf("failed to submit reinforcement quiz: %v", err)
	}

	// Create new assessment record for the reinforcement quiz
	if res != nil && res.Data.AIReview != "" {
		assessmentDomain := dto.BuildUserQuizAssessmentDomainForCreate(&dto.CreateUserQuizAssessmentRequest{
			UID:           req.UID,
			Questions:     req.Message,
			Answers:       answersStr,
			AIReview:      res.Data.AIReview,
			AIDetectGrade: res.Data.AIDetectGrade,
		})

		_, err := s.repo.Create(ctx, nil, assessmentDomain)
		if err != nil {
			logger.Errorf("Failed to save reinforcement quiz assessment for user %s: %v", req.UID, err)
		}
	}

	return status.SUCCESS, res, nil
}

func (s *userQuizAssessmentService) GetUserQuizAssessmentsHistory(ctx context.Context, req *dto.GetUserQuizAssessmentsHistoryRequest) (status.Code, *dto.UserQuizAssessmentsHistoryResponse, error) {
	logger := logger.GetLogger(ctx)

	assessments, paginationObj, err := s.repo.ListByUID(ctx, diRepo.ListUserQuizAssessmentsParams{
		UID:   req.UID,
		Page:  req.Page,
		Limit: req.Limit,
	})
	if err != nil {
		logger.Errorf("Failed to fetch quiz assessments history: %v", err)
		return status.FAIL, nil, fmt.Errorf("failed to fetch quiz assessments history: %v", err)
	}

	items := make([]dto.UserQuizAssessmentResponse, 0, len(assessments))
	for _, assessment := range assessments {
		items = append(items, dto.UserQuizAssessmentResponseFromDomain(assessment))
	}

	return status.SUCCESS, &dto.UserQuizAssessmentsHistoryResponse{
		Items:      items,
		Pagination: paginationObj,
	}, nil
}

func (s *userQuizAssessmentService) GetUserQuizAssessmentByID(ctx context.Context, id string) (status.Code, *dto.UserQuizAssessmentResponse, error) {
	logger := logger.GetLogger(ctx)
	assessment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.Errorf("Failed to fetch quiz assessment by ID: %v", err)
		return status.FAIL, nil, fmt.Errorf("failed to fetch quiz assessment by ID: %v", err)
	}

	if assessment == nil {
		logger.Errorf("Quiz assessment not found: %s", id)
		return status.NOT_FOUND, nil, fmt.Errorf("quiz assessment not found")
	}

	assessmentResp := dto.UserQuizAssessmentResponseFromDomain(assessment)
	return status.SUCCESS, &assessmentResp, nil
}

func (s *userQuizAssessmentService) CreateUserQuizAssessment(ctx context.Context, req *dto.CreateUserQuizAssessmentRequest) (status.Code, *dto.UserQuizAssessmentResponse, error) {
	logger := logger.GetLogger(ctx)

	uqaDomain := dto.BuildUserQuizAssessmentDomainForCreate(req)

	rowsAffected, err := s.repo.Create(ctx, nil, uqaDomain)
	if err != nil {
		logger.Errorf("Failed to create quiz assessment: %v", err)
		return status.FAIL, nil, fmt.Errorf("failed to create quiz assessment: %v", err)
	}

	if rowsAffected == 0 {
		return status.FAIL, nil, fmt.Errorf("failed to create user latest quiz")
	}

	response := dto.UserQuizAssessmentResponseFromDomain(uqaDomain)
	return status.SUCCESS, &response, nil
}

func (s *userQuizAssessmentService) UpdateUserQuizAssessment(ctx context.Context, req *dto.UpdateUserQuizAssessmentRequest) (status.Code, *dto.UserQuizAssessmentResponse, error) {
	logger := logger.GetLogger(ctx)

	existingQuiz, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.FAIL, nil, err
	}

	if existingQuiz == nil {
		return status.FAIL, nil, fmt.Errorf("user latest quiz not found")
	}

	uqaDomain := dto.BuildUserQuizAssessmentDomainForUpdate(req)

	rowsAffected, err := s.repo.Update(ctx, uqaDomain)
	if err != nil {
		logger.Errorf("Failed to update quiz assessment: %v", err)
		return status.FAIL, nil, fmt.Errorf("failed to update quiz assessment: %v", err)
	}

	if rowsAffected == 0 {
		return status.FAIL, nil, fmt.Errorf("failed to update user latest quiz")
	}

	updatedQuiz, err := s.repo.FindByID(ctx, existingQuiz.ID())
	if err != nil {
		return status.FAIL, nil, err
	}

	response := dto.UserQuizAssessmentResponseFromDomain(updatedQuiz)
	return status.SUCCESS, &response, nil
}

func (s *userQuizAssessmentService) DeleteUserQuizAssessment(ctx context.Context, req *dto.DeleteUserQuizAssessmentRequest) (status.Code, error) {
	existingQuiz, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return status.FAIL, err
	}

	if existingQuiz == nil {
		return status.FAIL, fmt.Errorf("user latest quiz not found")
	}

	rowsAffected, err := s.repo.Delete(ctx, req.ID)
	if err != nil {
		return status.FAIL, err
	}

	if rowsAffected == 0 {
		return status.FAIL, fmt.Errorf("failed to delete user latest quiz")
	}

	return status.SUCCESS, nil
}

func (s *userQuizAssessmentService) ForceDeleteUserQuizAssessment(ctx context.Context, req *dto.DeleteUserQuizAssessmentRequest) (status.Code, error) {
	logger := logger.GetLogger(ctx)

	rowsAffected, err := s.repo.ForceDelete(ctx, req.ID)
	if err != nil {
		logger.Errorf("Failed to force delete quiz assessment: %v", err)
		return status.FAIL, fmt.Errorf("failed to force delete quiz assessment: %v", err)
	}

	if rowsAffected == 0 {
		return status.FAIL, fmt.Errorf("failed to force delete user latest quiz")
	}

	return status.SUCCESS, nil
}
