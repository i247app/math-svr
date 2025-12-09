package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/applications/dto"
	helper "math-ai.com/math-ai/internal/applications/helpers/chatbox_helper"
	"math-ai.com/math-ai/internal/applications/validators"
	di "math-ai.com/math-ai/internal/core/di/services"
	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
	chatbox "math-ai.com/math-ai/internal/driven-adapter/external/chat-box"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
)

type ChatBoxService struct {
	validator            validators.IChatboxValidator
	client               chatbox.IChatBoxClient
	profileSvc           di.IProfileService
	userQuizPracticesSvc di.IUserQuizPracticesService
	jsonSanitizer        *helper.JSONSanitizer
}

func NewChatBoxService(
	client chatbox.IChatBoxClient,
	validator validators.IChatboxValidator,
) di.IChatBoxService {
	return &ChatBoxService{
		validator:     validator,
		client:        client,
		jsonSanitizer: helper.NewJSONSanitizer(),
	}
}

func (s *ChatBoxService) Generate(ctx context.Context, conv *domain.Conversation) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error) {
	// Send message to OpenAI
	resp, err := s.client.SendMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send message to OpenAI: %v", err)
		return status.FAIL, nil, fmt.Errorf("ChatBox service error: %v", err)
	}

	// Sanitize JSON response to fix common escaping issues using helper
	sanitizedJSON := s.jsonSanitizer.SanitizeJSONResponse(resp.Message)

	var data []dto.Question
	err = json.Unmarshal([]byte(sanitizedJSON), &data)
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

	return status.SUCCESS, response, nil
}

func (s *ChatBoxService) Submit(ctx context.Context, conv *domain.Conversation) (status.Code, *dto.ChatBoxResponse[dto.QuizAnswer], error) {
	// Send message to OpenAI
	resp, err := s.client.SendMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send message to OpenAI: %v", err)
		return status.FAIL, nil, fmt.Errorf("ChatBox service error: %v", err)
	}

	// Sanitize JSON response to fix common escaping issues using helper
	sanitizedJSON := s.jsonSanitizer.SanitizeJSONResponse(resp.Message)

	var data dto.QuizAnswer
	err = json.Unmarshal([]byte(sanitizedJSON), &data)
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

	return status.SUCCESS, response, nil
}

func (s *ChatBoxService) Reinforce(ctx context.Context, conv *domain.Conversation) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error) {
	// Send message to OpenAI
	resp, err := s.client.SendMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send message to OpenAI: %v", err)
		return status.FAIL, nil, fmt.Errorf("ChatBox service error: %v", err)
	}

	// Sanitize JSON response to fix common escaping issues using helper
	sanitizedJSON := s.jsonSanitizer.SanitizeJSONResponse(resp.Message)

	var data []dto.Question
	err = json.Unmarshal([]byte(sanitizedJSON), &data)
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

	return status.SUCCESS, response, nil
}
