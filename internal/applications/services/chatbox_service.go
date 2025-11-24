package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	di "math-ai.com/math-ai/internal/core/di/services"
	chatbox "math-ai.com/math-ai/internal/driven-adapter/external/chat-box"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
)

type ChatBoxService struct {
	validator         validators.IChatboxValidator
	client            chatbox.IChatBoxClient
	profileSvc        di.IProfileService
	userLatestQuizSvc di.IUserLatestQuizService
}

func NewChatBoxService(
	client chatbox.IChatBoxClient,
	validator validators.IChatboxValidator,
	profileSvc di.IProfileService,
	userLatestQuizSvc di.IUserLatestQuizService,
) di.IChatBoxService {
	return &ChatBoxService{
		validator:         validator,
		client:            client,
		profileSvc:        profileSvc,
		userLatestQuizSvc: userLatestQuizSvc,
	}
}

// sanitizeJSONResponse fixes common JSON escaping issues in AI responses,
// particularly for LaTeX expressions where backslashes need to be escaped.
func sanitizeJSONResponse(jsonStr string) string {
	// Pattern to find backslashes in JSON string values that are not already escaped
	// This regex looks for single backslashes followed by common LaTeX commands
	patterns := []struct {
		pattern string
		replace string
	}{
		// Fix unescaped backslashes before common LaTeX commands
		{`([^\\])\\frac`, `$1\\frac`},
		{`([^\\])\\sqrt`, `$1\\sqrt`},
		{`([^\\])\\int`, `$1\\int`},
		{`([^\\])\\{`, `$1\\{`},
		{`([^\\])\\}`, `$1\\}`},
		{`([^\\])\\left`, `$1\\left`},
		{`([^\\])\\right`, `$1\\right`},
		// Fix at start of string values (after quotes)
		{`"\\frac`, `"\\\\frac`},
		{`"\\sqrt`, `"\\\\sqrt`},
		{`"\\int`, `"\\\\int`},
		{`"\\{`, `"\\\\{`},
		{`"\\}`, `"\\\\}`},
		{`"\\left`, `"\\\\left`},
		{`"\\right`, `"\\\\right`},
	}

	result := jsonStr
	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		result = re.ReplaceAllString(result, p.replace)
	}

	return result
}

func (s *ChatBoxService) GenerateQuiz(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error) {
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

	// Send message to OpenAI
	resp, err := s.client.SendMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send message to OpenAI: %v", err)
		return status.INTERNAL, nil, fmt.Errorf("ChatBox service error: %v", err)
	}

	// Sanitize JSON response to fix common escaping issues
	sanitizedJSON := sanitizeJSONResponse(resp.Message)

	var data []dto.Question
	err = json.Unmarshal([]byte(sanitizedJSON), &data)
	if err != nil {
		logger.Errorf("Failed to unmarshal response message: %v", err)
		logger.Errorf("Original response: %s", resp.Message)
		logger.Errorf("Sanitized response: %s", sanitizedJSON)
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
			resetData := "?"
			_, _, err = s.userLatestQuizSvc.UpdateQuiz(ctx, &dto.UpdateUserLatestQuizRequest{
				// ID:        res.ID,
				UID:       res.UID,
				Questions: &resp.Message,
				Answers:   &resetData,
				AIReview:  &resetData,
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

	// Send message to OpenAI
	resp, err := s.client.SendMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send message to OpenAI: %v", err)
		return status.INTERNAL, nil, fmt.Errorf("ChatBox service error: %v", err)
	}

	// Sanitize JSON response to fix common escaping issues
	sanitizedJSON := sanitizeJSONResponse(resp.Message)

	var data dto.QuizAnswer
	err = json.Unmarshal([]byte(sanitizedJSON), &data)
	if err != nil {
		logger.Errorf("Failed to unmarshal response message: %v", err)
		logger.Errorf("Original response: %s", resp.Message)
		logger.Errorf("Sanitized response: %s", sanitizedJSON)
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
		// ID:       req.UserLatestQuizID,
		UID:      req.UID,
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

func (s *ChatBoxService) GenerateQuizPractice(ctx context.Context, req *dto.GenerateQuizPracticeRequest) (status.Code, *dto.ChatBoxResponse[[]dto.Question], error) {
	statusCode, ulq, err := s.userLatestQuizSvc.GetQuizByUID(ctx, &dto.GetUserLatestQuizByUIDRequest{
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

	// Send message to OpenAI
	resp, err := s.client.SendMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send message to OpenAI: %v", err)
		return status.INTERNAL, nil, fmt.Errorf("ChatBox service error: %v", err)
	}

	// Sanitize JSON response to fix common escaping issues
	sanitizedJSON := sanitizeJSONResponse(resp.Message)

	var data []dto.Question
	err = json.Unmarshal([]byte(sanitizedJSON), &data)
	if err != nil {
		logger.Errorf("Failed to unmarshal response message: %v", err)
		logger.Errorf("Original response: %s", resp.Message)
		logger.Errorf("Sanitized response: %s", sanitizedJSON)
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
		resetData := "?"
		_, _, err := s.userLatestQuizSvc.UpdateQuiz(ctx, &dto.UpdateUserLatestQuizRequest{
			ID:        ulq.ID,
			UID:       ulq.UID,
			Questions: &resp.Message,
			Answers:   &resetData,
			AIReview:  &resetData,
		})
		if err != nil {
			logger.Errorf("Failed to update latest quiz for user %s: %v", req.UID, err)
		}
	}

	// Include conversation history if requested
	if req.History != nil {
		response.History = dto.ConversationToHistoryDTO(conv)
	}

	return status.SUCCESS, response, nil
}

func (s *ChatBoxService) SendMessageStream(ctx context.Context, req *dto.GenerateQuizRequest) (status.Code, <-chan dto.ChatBoxStreamChunk, error) {
	// Build conversation from request
	conv := dto.BuildGenerateQuizFromRequest(ctx, req, nil)

	// Send message to OpenAI with streaming
	streamChan, err := s.client.StreamMessage(ctx, conv)
	if err != nil {
		logger.Errorf("Failed to send streaming message to OpenAI: %v", err)
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
