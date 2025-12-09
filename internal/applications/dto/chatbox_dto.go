package dto

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	appctx "math-ai.com/math-ai/internal/shared/utils/context"
)

// MessageDTO represents a message in the conversation
type MessageDTO struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

type ChatBoxRequestCommon struct {
	Message      string        `json:"message" binding:"required"`
	History      []*MessageDTO `json:"history,omitempty"`
	Model        *string       `json:"model,omitempty"`
	Temperature  *float32      `json:"temperature,omitempty"`
	MaxTokens    *int          `json:"max_tokens,omitempty"`
	SystemPrompt *string       `json:"system_prompt,omitempty"`
	Stream       bool          `json:"stream,omitempty"`
}

// ChatBoxRequest represents a request to send a message to the chatbox
type GenerateQuizRequest struct {
	UID           string                `json:"uid"`
	TypeOfQuiz    string                `json:"type_of_task,omitempty"`
	TypeOfPurpose enum.ETypeQuizPurpuse `json:"type_of_purpose,omitempty"`
	ChatBoxRequestCommon
}

type Question struct {
	QuestionNumber int64  `json:"question_number"`
	QuestionName   string `json:"question_name"`
	Answers        []struct {
		Label   string `json:"label"`
		Content string `json:"content"`
	} `json:"answers"`
	RightAnswer string `json:"right_answer"`
}

// ChatBoxResponse represents the response from the chatbox
type ChatBoxResponse[T any] struct {
	Response         string        `json:"response"`
	Data             T             `json:"data"`
	Role             string        `json:"role"`
	Model            string        `json:"model"`
	FinishReason     string        `json:"-"`
	PromptTokens     int           `json:"-"`
	CompletionTokens int           `json:"-"`
	TotalTokens      int           `json:"-"`
	History          []*MessageDTO `json:"-"`
	Timestamp        time.Time     `json:"timestamp"`
}

type GenerateQuizResponse struct {
	Result *ChatBoxResponse[[]Question] `json:"result"`
}

type SubmitQuizRequest struct {
	// UserLatestQuizID string `json:"user_latest_quiz_id"`
	UID     string `json:"uid"`
	Answers []struct {
		QuestionNumber int64  `json:"question_number"`
		Answer         string `json:"answer"`
	} `json:"answers"`
	ChatBoxRequestCommon
}

type QuizAnswer struct {
	TotalQuestions  int64  `json:"total_questions"`
	CorrectNumber   int64  `json:"correct_number"`
	ScorePercentage int    `json:"score_percentage"`
	AIReview        string `json:"ai_review"`
}

type SubmitQuizResponse struct {
	Result *ChatBoxResponse[QuizAnswer] `json:"result"`
}

type GenerateQuizPracticeRequest struct {
	UID string `json:"uid"`
	ChatBoxRequestCommon
}

type GenerateQuizPracticeResponse struct {
	Result *ChatBoxResponse[[]Question] `json:"result"`
}

// ChatBoxStreamChunk represents a chunk in a streaming response
type ChatBoxStreamChunk struct {
	Delta        string `json:"delta"`
	FinishReason string `json:"finish_reason,omitempty"`
	Done         bool   `json:"done"`
	Error        error  `json:"error,omitempty"`
}

func BuildGenerateQuizFromRequest(ctx context.Context, req *GenerateQuizRequest, userProfile *ProfileResponse) *domain.Conversation {
	var (
		language string
		grade    string
		semester string
	)

	switch appctx.GetLocale(ctx) {
	case "en":
		language = "English"
	case "vn":
		language = "Vietnamese"
	default:
		language = "English"
	}

	if userProfile != nil {
		grade = userProfile.Grade
		semester = userProfile.Semester
	}

	conv := domain.NewConversation()

	// Set model if provided
	if req.Model != nil {
		conv.SetModel(*req.Model)
	}

	// Set temperature if provided
	if req.Temperature != nil {
		conv.SetTemperature(*req.Temperature)
	}

	// Set max tokens if provided
	if req.MaxTokens != nil {
		conv.SetMaxTokens(*req.MaxTokens)
	}

	// Set system prompt if provided
	if req.SystemPrompt != nil {
		conv.SetSystemPrompt(req.SystemPrompt)
	}

	prompt := fmt.Sprintf(domain.PromptMathQuizNew, grade, semester, language)

	// Add the current user message
	userMsg := domain.NewMessage("user", prompt)
	conv.AddMessage(userMsg)

	return conv
}

func BuildSubmitQuizAnswerFromRequest(ctx context.Context, req *SubmitQuizRequest, userQuizPractice *UserQuizPracticesResponse) *domain.Conversation {
	var (
		language             string
		questionsInformation string
		userAnswers          string
	)

	switch appctx.GetLocale(ctx) {
	case "en":
		language = "English"
	case "vn":
		language = "Vietnamese"
	default:
		language = "English"
	}

	if req != nil {
		questionsInformation = userQuizPractice.Questions
		userAnswers = userQuizPractice.Answers
	}

	conv := domain.NewConversation()

	// Set model if provided
	if req.Model != nil {
		conv.SetModel(*req.Model)
	}

	// Set temperature if provided
	if req.Temperature != nil {
		conv.SetTemperature(*req.Temperature)
	}

	// Set max tokens if provided
	if req.MaxTokens != nil {
		conv.SetMaxTokens(*req.MaxTokens)
	}

	// Set system prompt if provided
	if req.SystemPrompt != nil {
		conv.SetSystemPrompt(req.SystemPrompt)
	}

	prompt := fmt.Sprintf(domain.SubmitQuizAnswerPrompt, questionsInformation, userAnswers, language)

	// Add the current user message
	userMsg := domain.NewMessage("user", prompt)
	conv.AddMessage(userMsg)

	return conv
}

func BuildGeneratePracticeQuizFromRequest(ctx context.Context, req *GenerateQuizPracticeRequest, userQuizPractice *UserQuizPracticesResponse) *domain.Conversation {
	var (
		language             string
		questionsInformation string
		userAnswers          string
		reviewedPerformance  string
	)

	switch appctx.GetLocale(ctx) {
	case "en":
		language = "English"
	case "vn":
		language = "Vietnamese"
	default:
		language = "English"
	}

	if req != nil {
		questionsInformation = userQuizPractice.Questions
		userAnswers = userQuizPractice.Answers
		reviewedPerformance = userQuizPractice.AIReview
	}

	conv := domain.NewConversation()

	// Set model if provided
	if req.Model != nil {
		conv.SetModel(*req.Model)
	}

	// Set temperature if provided
	if req.Temperature != nil {
		conv.SetTemperature(*req.Temperature)
	}

	// Set max tokens if provided
	if req.MaxTokens != nil {
		conv.SetMaxTokens(*req.MaxTokens)
	}

	// Set system prompt if provided
	if req.SystemPrompt != nil {
		conv.SetSystemPrompt(req.SystemPrompt)
	}

	prompt := fmt.Sprintf(domain.PromptMathQuizPractice, questionsInformation, userAnswers, reviewedPerformance, language)

	// Add the current user message
	userMsg := domain.NewMessage("user", prompt)
	conv.AddMessage(userMsg)

	return conv
}

func MessageDomainToDTO(msg *domain.Message) *MessageDTO {
	return &MessageDTO{
		Role:      msg.Role(),
		Content:   msg.Content(),
		Timestamp: msg.Timestamp(),
	}
}

// BuildSubmitQuizAnswerForAssessment builds a conversation for submitting quiz answers with grade assessment
func BuildSubmitQuizAnswerForAssessment(ctx context.Context, req *SubmitQuizAssessmentRequest, currentGrade string) *domain.Conversation {
	var (
		language             string
		questionsInformation string
		userAnswers          string
	)

	switch appctx.GetLocale(ctx) {
	case "en":
		language = "English"
	case "vn":
		language = "Vietnamese"
	default:
		language = "English"
	}

	if req != nil {
		questionsInformation = req.Message
		answersJSON, _ := json.Marshal(req.Answers)
		userAnswers = string(answersJSON)
	}

	conv := domain.NewConversation()

	// Set model if provided
	if req.Model != nil {
		conv.SetModel(*req.Model)
	}

	// Set temperature if provided
	if req.Temperature != nil {
		conv.SetTemperature(*req.Temperature)
	}

	// Set max tokens if provided
	if req.MaxTokens != nil {
		conv.SetMaxTokens(*req.MaxTokens)
	}

	// Set system prompt if provided
	if req.SystemPrompt != nil {
		conv.SetSystemPrompt(req.SystemPrompt)
	}

	prompt := fmt.Sprintf(domain.SubmitQuizAnswerAssessmentPrompt, questionsInformation, userAnswers, currentGrade, language)

	// Add the current user message
	userMsg := domain.NewMessage("user", prompt)
	conv.AddMessage(userMsg)

	return conv
}

// BuildReinforceQuizAssessmentFromRequest builds a conversation for generating reinforcement quiz
func BuildReinforceQuizAssessmentFromRequest(ctx context.Context, req *ReinforceQuizAssessmentRequest, assessment *UserQuizAssessmentResponse) *domain.Conversation {
	var (
		language             string
		questionsInformation string
		userAnswers          string
		reviewedPerformance  string
	)

	switch appctx.GetLocale(ctx) {
	case "en":
		language = "English"
	case "vn":
		language = "Vietnamese"
	default:
		language = "English"
	}

	if assessment != nil {
		questionsInformation = assessment.Questions
		userAnswers = assessment.Answers
		reviewedPerformance = assessment.AIReview
	}

	conv := domain.NewConversation()

	// Set model if provided
	if req.Model != nil {
		conv.SetModel(*req.Model)
	}

	// Set temperature if provided
	if req.Temperature != nil {
		conv.SetTemperature(*req.Temperature)
	}

	// Set max tokens if provided
	if req.MaxTokens != nil {
		conv.SetMaxTokens(*req.MaxTokens)
	}

	// Set system prompt if provided
	if req.SystemPrompt != nil {
		conv.SetSystemPrompt(req.SystemPrompt)
	}

	prompt := fmt.Sprintf(domain.PromptMathQuizPractice, questionsInformation, userAnswers, reviewedPerformance, language)

	// Add the current user message
	userMsg := domain.NewMessage("user", prompt)
	conv.AddMessage(userMsg)

	return conv
}
