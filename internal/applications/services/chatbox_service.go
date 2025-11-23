package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	chatbox "math-ai.com/math-ai/internal/driven-adapter/external/chat-box"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
)

type ChatBoxService struct {
	client            chatbox.IChatBoxClient
	profileSvc        di.IProfileService
	userLatestQuizSvc di.IUserLatestQuizService
}

func NewChatBoxService(
	client chatbox.IChatBoxClient,
	profileSvc di.IProfileService,
	userLatestQuizSvc di.IUserLatestQuizService,
) di.IChatBoxService {
	return &ChatBoxService{
		client:            client,
		profileSvc:        profileSvc,
		userLatestQuizSvc: userLatestQuizSvc,
	}
}

func (s *ChatBoxService) GenerateQuiz(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error) {
	statusCode, user, err := s.profileSvc.FetchProfile(ctx, &dto.FetchProfileRequest{
		UID: req.UID, // Replace with actual user UID from context or request
	})
	if err != nil {
		logger.Errorf("Failed to fetch user profile: %v", err)
		return statusCode, nil, fmt.Errorf("failed to fetch user profile: %v", err)
	}

	// Build generate quiz from request
	conv := dto.BuildGenerateQuizFromRequest(req, user)

	// log prompt for debugging
	for _, msg := range conv.Messages() {
		if msg.Role() == "user" {
			logger.Infof("User prompt: %s", msg.Content())
		}
	}

	// Send message to OpenAI
	resp, err := s.client.SendMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send message to OpenAI: %v", err)
		// Check for specific OpenAI errors and return appropriate status codes
		errMsg := err.Error()
		if contains(errMsg, "status code: 429") || contains(errMsg, "exceeded your current quota") {
			return status.INTERNAL, nil, fmt.Errorf("OpenAI API quota exceeded. Please check your billing details at https://platform.openai.com/account/billing")
		}
		if contains(errMsg, "status code: 401") || contains(errMsg, "invalid api key") {
			return status.INTERNAL, nil, fmt.Errorf("Invalid OpenAI API key. Please check your CHAT_BOX_API_KEY configuration")
		}
		return status.INTERNAL, nil, fmt.Errorf("ChatBox service error: %v", err)
	}

	var data []dto.Question
	err = json.Unmarshal([]byte(resp.Message), &data)
	if err != nil {
		logger.Errorf("Failed to unmarshal response message: %v", err)
	}

	// Build response DTO
	response := &dto.ChatBoxResponse[[]dto.Question]{
		Response:         resp.Message,
		Data:             data,
		Role:             resp.Role,
		Model:            resp.Model,
		FinishReason:     resp.FinishReason,
		PromptTokens:     resp.PromptTokens,
		CompletionTokens: resp.CompletionTokens,
		TotalTokens:      resp.TotalTokens,
		Timestamp:        time.Now(),
	}

	// Save latest quiz for the user
	if resp.Message != "" {
		_, res, err := s.userLatestQuizSvc.GetQuizByUID(ctx, &dto.GetUserLatestQuizByUIDRequest{
			UID: req.UID,
		})
		if err != nil {
			logger.Errorf("Failed to get latest quiz for user %s: %v", req.UID, err)
		}

		if res == nil {
			_, createdRes, err := s.userLatestQuizSvc.CreateQuiz(ctx, &dto.CreateUserLatestQuizRequest{
				UID:       req.UID,
				Questions: resp.Message,
				AIReview:  "",
			})
			if err != nil {
				logger.Errorf("Failed to create latest quiz for user %s: %v", req.UID, err)
			}
			response.UserLatesQuizID = createdRes.ID
		} else {
			_, _, err = s.userLatestQuizSvc.UpdateQuiz(ctx, &dto.UpdateUserLatestQuizRequest{
				ID:        res.ID,
				Questions: &resp.Message,
			})
			if err != nil {
				logger.Errorf("Failed to update latest quiz for user %s: %v", req.UID, err)
			}
			response.UserLatesQuizID = res.ID
		}
	}

	// Include conversation history if requested
	if req.History != nil {
		response.History = dto.ConversationToHistoryDTO(conv)
	}

	return status.SUCCESS, response, nil
}

func (s *ChatBoxService) SubmitQuiz(ctx context.Context, req *dto.SubmitQuizRequest) (status.Code, *dto.ChatBoxResponse[dto.QuizAnswer], error) {
	jsonAnswers, err := json.Marshal(req.Answers)
	if err != nil {
		log.Fatalf("Error marshaling struct to JSON: %v", err)
	}

	answersStr := string(jsonAnswers)

	statusCode, ulq, err := s.userLatestQuizSvc.UpdateQuiz(ctx, &dto.UpdateUserLatestQuizRequest{
		ID:      req.UserLatestQuizID,
		Answers: &answersStr,
	})

	if err != nil {
		logger.Errorf("Failed to fetch user profile: %v", err)
		return statusCode, nil, fmt.Errorf("failed to fetch user profile: %v", err)
	}

	// Build generate quiz from request
	conv := dto.BuildSubmitQuizAnswerFromRequest(req, ulq)

	// log prompt for debugging
	for _, msg := range conv.Messages() {
		if msg.Role() == "user" {
			logger.Infof("User prompt: %s", msg.Content())
		}
	}

	// Send message to OpenAI
	resp, err := s.client.SendMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send message to OpenAI: %v", err)
		// Check for specific OpenAI errors and return appropriate status codes
		errMsg := err.Error()
		if contains(errMsg, "status code: 429") || contains(errMsg, "exceeded your current quota") {
			return status.INTERNAL, nil, fmt.Errorf("OpenAI API quota exceeded. Please check your billing details at https://platform.openai.com/account/billing")
		}
		if contains(errMsg, "status code: 401") || contains(errMsg, "invalid api key") {
			return status.INTERNAL, nil, fmt.Errorf("Invalid OpenAI API key. Please check your CHAT_BOX_API_KEY configuration")
		}
		return status.INTERNAL, nil, fmt.Errorf("ChatBox service error: %v", err)
	}

	var data dto.QuizAnswer
	err = json.Unmarshal([]byte(resp.Message), &data)
	if err != nil {
		logger.Errorf("Failed to unmarshal response message: %v", err)
	}

	// Build response DTO
	response := &dto.ChatBoxResponse[dto.QuizAnswer]{
		Response:         resp.Message,
		Data:             data,
		Role:             resp.Role,
		Model:            resp.Model,
		FinishReason:     resp.FinishReason,
		PromptTokens:     resp.PromptTokens,
		CompletionTokens: resp.CompletionTokens,
		TotalTokens:      resp.TotalTokens,
		Timestamp:        time.Now(),
	}

	statusCode, _, err = s.userLatestQuizSvc.UpdateQuiz(ctx, &dto.UpdateUserLatestQuizRequest{
		ID:       req.UserLatestQuizID,
		AIReview: &data.AIReview,
	})
	if err != nil {
		logger.Errorf("Failed to update user latest quiz with AI review: %v", err)
		return statusCode, nil, fmt.Errorf("failed to update user latest quiz with AI review: %v", err)
	}

	// Include conversation history if requested
	if req.History != nil {
		response.History = dto.ConversationToHistoryDTO(conv)
	}

	return status.SUCCESS, response, nil
}

func (s *ChatBoxService) SendMessageStream(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, <-chan dto.ChatBoxStreamChunk, error) {
	// Build conversation from request
	conv := dto.BuildGenerateQuizFromRequest(req, nil)

	// Send message to OpenAI with streaming
	streamChan, err := s.client.StreamMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send streaming message to OpenAI: %v", err)
		// Check for specific OpenAI errors
		errMsg := err.Error()
		if contains(errMsg, "status code: 429") || contains(errMsg, "exceeded your current quota") {
			return status.INTERNAL, nil, fmt.Errorf("OpenAI API quota exceeded. Please check your billing details at https://platform.openai.com/account/billing")
		}
		if contains(errMsg, "status code: 401") || contains(errMsg, "invalid api key") {
			return status.INTERNAL, nil, fmt.Errorf("Invalid OpenAI API key. Please check your CHAT_BOX_API_KEY configuration")
		}
		return status.INTERNAL, nil, fmt.Errorf("ChatBox service error: %v", err)
	}

	// Create output channel
	outputChan := make(chan dto.ChatBoxStreamChunk)

	// Start goroutine to convert stream chunks to DTOs
	go func() {
		defer close(outputChan)

		for chunk := range streamChan {
			outputChan <- dto.ChatBoxStreamChunk{
				Delta:        chunk.Delta,
				FinishReason: chunk.FinishReason,
				Done:         chunk.Done,
				Error:        chunk.Error,
			}
		}
	}()

	return status.SUCCESS, outputChan, nil
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
