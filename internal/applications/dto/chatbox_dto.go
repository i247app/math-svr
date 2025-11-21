package dto

import (
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/chatbox"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

// MessageDTO represents a message in the conversation
type MessageDTO struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// ChatBoxRequest represents a request to send a message to the chatbox
type ChatBoxRequest struct {
	Message       string                `json:"message" binding:"required"`
	History       []*MessageDTO         `json:"history,omitempty"`
	Model         *string               `json:"model,omitempty"`
	Temperature   *float32              `json:"temperature,omitempty"`
	MaxTokens     *int                  `json:"max_tokens,omitempty"`
	SystemPrompt  *string               `json:"system_prompt,omitempty"`
	Stream        bool                  `json:"stream,omitempty"`
	TypeOfQuiz    string                `json:"type_of_task,omitempty"`
	TypeOfPurpose enum.ETypeQuizPurpuse `json:"type_of_purpose,omitempty"`
}

type Data struct {
	Question string `json:"question"`
	Answer   []struct {
		Label   string `json:"label"`
		Content string `json:"content"`
	} `json:"answers"`
	RightAnswer string `json:"right_answer"`
	Duration    int    `json:"duration"`
}

// ChatBoxResponse represents the response from the chatbox
type ChatBoxResponse struct {
	Response         string        `json:"response"`
	Data             []Data        `json:"data,omitempty"`
	Role             string        `json:"role"`
	Model            string        `json:"model"`
	FinishReason     string        `json:"finish_reason,omitempty"`
	PromptTokens     int           `json:"prompt_tokens,omitempty"`
	CompletionTokens int           `json:"completion_tokens,omitempty"`
	TotalTokens      int           `json:"total_tokens,omitempty"`
	History          []*MessageDTO `json:"history,omitempty"`
	Timestamp        time.Time     `json:"timestamp"`
}

type AskChatBoxResponse struct {
	Result *ChatBoxResponse `json:"result"`
}

// ChatBoxStreamChunk represents a chunk in a streaming response
type ChatBoxStreamChunk struct {
	Delta        string `json:"delta"`
	FinishReason string `json:"finish_reason,omitempty"`
	Done         bool   `json:"done"`
	Error        error  `json:"error,omitempty"`
}

// BuildConversationFromRequest builds a Conversation domain object from a ChatBoxRequest
func BuildConversationFromRequest(req *ChatBoxRequest) *domain.Conversation {
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

	// Add history messages
	if req.History != nil && len(req.History) > 0 {
		for _, msgDTO := range req.History {
			msg := domain.NewMessage(msgDTO.Role, msgDTO.Content)
			if !msgDTO.Timestamp.IsZero() {
				msg.SetTimestamp(msgDTO.Timestamp)
			}
			conv.AddMessage(msg)
		}
	}

	var prompt string
	switch req.TypeOfPurpose {
	case enum.TypeQuizPurpuseNew:
		prompt = domain.PromptMathQuizNew
	case enum.TypeQuizPurpusePractice:
		prompt = domain.PromptMathQuizPractice
	case enum.TypeQuizPurpuseExam:
		prompt = domain.PromptMathQuizExam
	default:
		prompt = domain.PromptMathQuizNew
	}

	// Add the current user message
	userMsg := domain.NewMessage("user", prompt)
	conv.AddMessage(userMsg)

	return conv
}

// MessageDomainToDTO converts a domain Message to a MessageDTO
func MessageDomainToDTO(msg *domain.Message) *MessageDTO {
	return &MessageDTO{
		Role:      msg.Role(),
		Content:   msg.Content(),
		Timestamp: msg.Timestamp(),
	}
}

// ConversationToHistoryDTO converts conversation messages to MessageDTOs
func ConversationToHistoryDTO(conv *domain.Conversation) []*MessageDTO {
	messages := conv.Messages()
	history := make([]*MessageDTO, len(messages))

	for i, msg := range messages {
		history[i] = MessageDomainToDTO(msg)
	}

	return history
}
