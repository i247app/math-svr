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
	conv := dto.BuildChatDomainForGenerateQuizPractice(ctx, &dto.GenerateQuizRequest{
		UID:                  req.UID,
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

	return status.SUCCESS, res, nil
}

func (s *userQuizAssessmentService) SubmitQuizAssessment(ctx context.Context, req *dto.SubmitQuizAssessmentRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAssessmentAnswer], error) {
	logger := logger.GetLogger(ctx)

	// Get user profile to determine current grade
	statusCode, user, err := s.profileSvc.FetchProfile(ctx, &dto.FetchProfileRequest{
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
	conv := dto.BuildChatDomainSubmitQuizAssessment(ctx, req, user.Grade)

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
		assessmentDomain := dto.BuildUserQuizAssessmentDomainForCreate(
			req.UID,
			req.Message, // questions
			answersStr,  // answers
			res.Data.AIReview,
			res.Data.AIDetectGrade,
		)

		_, err := s.repo.Create(ctx, nil, assessmentDomain)
		if err != nil {
			logger.Errorf("Failed to save quiz assessment for user %s: %v", req.UID, err)
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
	statusCode, assessment, err := s.GetUserQuizPraticeByID(ctx, req.UserQuizAssessmentID)
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
		assessmentDomain := dto.BuildUserQuizAssessmentDomainForCreate(
			req.UID,
			req.Message, // questions
			answersStr,  // answers
			res.Data.AIReview,
			res.Data.AIDetectGrade,
		)

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

func (s *userQuizAssessmentService) GetUserQuizPraticeByID(ctx context.Context, id string) (status.Code, *dto.UserQuizAssessmentResponse, error) {
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
