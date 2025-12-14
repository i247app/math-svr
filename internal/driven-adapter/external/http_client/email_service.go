package http_client

import (
	"context"
	"fmt"
	"time"
)

// EmailService handles email-related HTTP requests
type EmailService struct {
	client *Client
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

// NewEmailService creates a new email service with configurable options
// Example usage:
//   service := NewEmailService(
//     WithBaseURL("https://api.emailprovider.com"),
//     WithEmailKey("your-api-key"),
//     WithTimeout(30 * time.Second),
//   )
func NewEmailService(opts ...Option) *EmailService {
	// Add default options for email service
	defaultOpts := []Option{
		WithContentType("application/json"),
		WithAccept("application/json"),
		WithUserAgent("Math-AI-Email-Client/1.0"),
		WithTimeout(30 * time.Second),
	}

	// Merge default options with provided options
	allOpts := append(defaultOpts, opts...)

	return &EmailService{
		client: NewClient(allOpts...),
	}
}

// SendEmail sends an email via the configured provider
func (s *EmailService) SendEmail(ctx context.Context, req *EmailRequest) (*EmailResponse, error) {
	resp, err := s.client.Post(ctx, "/send", req)
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("email API error: %s", resp.String())
	}

	var emailResp EmailResponse
	if err := resp.JSON(&emailResp); err != nil {
		return nil, fmt.Errorf("failed to parse email response: %w", err)
	}

	return &emailResp, nil
}

// SendBulkEmail sends multiple emails
func (s *EmailService) SendBulkEmail(ctx context.Context, emails []*EmailRequest) error {
	resp, err := s.client.Post(ctx, "/send/bulk", map[string]interface{}{
		"emails": emails,
	})
	if err != nil {
		return fmt.Errorf("failed to send bulk emails: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("bulk email API error: %s", resp.String())
	}

	return nil
}

// GetEmailStatus checks the status of a sent email
func (s *EmailService) GetEmailStatus(ctx context.Context, messageID string) (string, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/status/%s", messageID))
	if err != nil {
		return "", fmt.Errorf("failed to get email status: %w", err)
	}

	if !resp.IsSuccess() {
		return "", fmt.Errorf("email status API error: %s", resp.String())
	}

	var statusResp struct {
		Status string `json:"status"`
	}
	if err := resp.JSON(&statusResp); err != nil {
		return "", fmt.Errorf("failed to parse status response: %w", err)
	}

	return statusResp.Status, nil
}
