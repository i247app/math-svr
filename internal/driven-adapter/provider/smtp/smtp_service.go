package smtp

import (
	"context"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/driven-adapter/external/http_client"
)

// SMTPService handles SMTP-based email-related HTTP requests
type SMTPService struct {
	client *http_client.Client
}

func NewSMTPService(opts ...http_client.Option) *SMTPService {
	// Add default options for SMTP service
	defaultOpts := []http_client.Option{
		http_client.WithContentType("application/json"),
		http_client.WithAccept("application/json"),
		http_client.WithUserAgent("Math-AI-SMTP-Client/1.0"),
		http_client.WithTimeout(60 * time.Second), // Longer timeout for email with attachments
	}

	// Merge default options with provided options
	allOpts := append(defaultOpts, opts...)

	return &SMTPService{
		client: http_client.NewClient(allOpts...),
	}
}

// SendEmail sends an email via SMTP provider
func (s *SMTPService) SendEmail(ctx context.Context, req *SMTPEmailRequest) (*SMTPEmailResponse, error) {
	resp, err := s.client.Post(ctx, "/v1/send", req)
	if err != nil {
		return nil, fmt.Errorf("failed to send SMTP email: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("SMTP API error: %s", resp.String())
	}

	var emailResp SMTPEmailResponse
	if err := resp.JSON(&emailResp); err != nil {
		return nil, fmt.Errorf("failed to parse SMTP response: %w", err)
	}

	return &emailResp, nil
}

// SendTemplate sends an email using a template
func (s *SMTPService) SendTemplate(ctx context.Context, templateID string, to []string, variables map[string]interface{}) (*SMTPEmailResponse, error) {
	payload := map[string]interface{}{
		"template_id": templateID,
		"to":          to,
		"variables":   variables,
	}

	resp, err := s.client.Post(ctx, "/v1/send/template", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to send template email: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("SMTP template API error: %s", resp.String())
	}

	var emailResp SMTPEmailResponse
	if err := resp.JSON(&emailResp); err != nil {
		return nil, fmt.Errorf("failed to parse template response: %w", err)
	}

	return &emailResp, nil
}

// VerifyEmail verifies if an email address is valid
func (s *SMTPService) VerifyEmail(ctx context.Context, email string) (bool, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/v1/verify/%s", email))
	if err != nil {
		return false, fmt.Errorf("failed to verify email: %w", err)
	}

	if !resp.IsSuccess() {
		return false, fmt.Errorf("SMTP verify API error: %s", resp.String())
	}

	var verifyResp struct {
		Valid bool `json:"valid"`
	}
	if err := resp.JSON(&verifyResp); err != nil {
		return false, fmt.Errorf("failed to parse verify response: %w", err)
	}

	return verifyResp.Valid, nil
}
