package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	chatbox "math-ai.com/math-ai/internal/driven-adapter/external/chat-box"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
)

type ChatBoxService struct {
	client     chatbox.IChatBoxClient
	profileSvc di.IProfileService
}

func NewChatBoxService(client chatbox.IChatBoxClient, profileSvc di.IProfileService) di.IChatBoxService {
	return &ChatBoxService{
		client:     client,
		profileSvc: profileSvc,
	}
}

// SendMessage sends a message to the chatbox and gets a response
func (s *ChatBoxService) SendMessage(ctx context.Context, req *dto.ChatBoxRequest) (status.Code, *dto.ChatBoxResponse, error) {
	statusCode, user, err := s.profileSvc.FetchProfile(ctx, &dto.FetchProfileRequest{
		UID: req.UID, // Replace with actual user UID from context or request
	})
	if err != nil {
		logger.Errorf("Failed to fetch user profile: %v", err)
		return statusCode, nil, fmt.Errorf("failed to fetch user profile: %v", err)
	}

	// Build conversation from request
	conv := dto.BuildConversationFromRequest(req, user)

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

	// Build response DTO
	response := &dto.ChatBoxResponse{
		Response:         resp.Message,
		Role:             resp.Role,
		Model:            resp.Model,
		FinishReason:     resp.FinishReason,
		PromptTokens:     resp.PromptTokens,
		CompletionTokens: resp.CompletionTokens,
		TotalTokens:      resp.TotalTokens,
		Timestamp:        time.Now(),
	}

	err = json.Unmarshal([]byte(resp.Message), &response.Data)
	if err != nil {
		logger.Errorf("Failed to unmarshal response message: %v", err)
		// return status.INTERNAL, nil, fmt.Errorf("Failed to parse chatbox response: %v", err)
	}

	// Include conversation history if requested
	if req.History != nil {
		response.History = dto.ConversationToHistoryDTO(conv)
	}

	return status.SUCCESS, response, nil
}

// SendMessageStream sends a message and streams the response
func (s *ChatBoxService) SendMessageStream(ctx context.Context, req *dto.ChatBoxRequest) (status.Code, <-chan dto.ChatBoxStreamChunk, error) {
	// Build conversation from request
	conv := dto.BuildConversationFromRequest(req, nil)

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

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
