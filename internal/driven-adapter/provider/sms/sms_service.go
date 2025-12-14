package http_client

import (
	"context"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/driven-adapter/external/http_client"
)

// SmsService handles SMS-related HTTP requests
type SmsService struct {
	client *http_client.Client
}

func NewSmsService(opts ...http_client.Option) *SmsService {
	// Add default options for SMS service
	defaultOpts := []http_client.Option{
		http_client.WithContentType("application/json"),
		http_client.WithAccept("application/json"),
		http_client.WithUserAgent("Math-AI-SMS-Client/1.0"),
		http_client.WithTimeout(30 * time.Second),
	}

	// Merge default options with provided options
	allOpts := append(defaultOpts, opts...)

	return &SmsService{
		client: http_client.NewClient(allOpts...),
	}
}

// SendSMS sends an SMS message via the configured provider
func (s *SmsService) SendSMS(ctx context.Context, req *SMSRequest) (*SMSResponse, error) {
	resp, err := s.client.Post(ctx, "/messages", req)
	if err != nil {
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("SMS API error: %s", resp.String())
	}

	var smsResp SMSResponse
	if err := resp.JSON(&smsResp); err != nil {
		return nil, fmt.Errorf("failed to parse SMS response: %w", err)
	}

	return &smsResp, nil
}

// SendBulkSMS sends multiple SMS messages
func (s *SmsService) SendBulkSMS(ctx context.Context, messages []*SMSRequest) ([]*SMSResponse, error) {
	resp, err := s.client.Post(ctx, "/messages/bulk", map[string]interface{}{
		"messages": messages,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send bulk SMS: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("bulk SMS API error: %s", resp.String())
	}

	var responses []*SMSResponse
	if err := resp.JSON(&responses); err != nil {
		return nil, fmt.Errorf("failed to parse bulk SMS response: %w", err)
	}

	return responses, nil
}

// GetSMSStatus checks the status of a sent SMS
func (s *SmsService) GetSMSStatus(ctx context.Context, messageID string) (string, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/messages/%s", messageID))
	if err != nil {
		return "", fmt.Errorf("failed to get SMS status: %w", err)
	}

	if !resp.IsSuccess() {
		return "", fmt.Errorf("SMS status API error: %s", resp.String())
	}

	var statusResp struct {
		Status string `json:"status"`
	}
	if err := resp.JSON(&statusResp); err != nil {
		return "", fmt.Errorf("failed to parse status response: %w", err)
	}

	return statusResp.Status, nil
}
