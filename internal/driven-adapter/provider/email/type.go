package email

import "math-ai.com/math-ai/internal/driven-adapter/external/http_client"

// EmailService handles email-related HTTP requests
type EmailService struct {
	client *http_client.Client
}

// EmailRequest represents an email request payload
type EmailRequest struct {
	To      []string `json:"to"`
	From    string   `json:"from"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	HTML    string   `json:"html,omitempty"`
}

// EmailResponse represents an email response
type EmailResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
}
