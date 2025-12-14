package http_client

import (
	"context"
	"fmt"
	"time"
)

// PushNotificationService handles push notification-related HTTP requests
type PushNotificationService struct {
	client *Client
}

// PushNotificationRequest represents a push notification request payload
type PushNotificationRequest struct {
	To    []string               `json:"to"`
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data,omitempty"`
	Badge int                    `json:"badge,omitempty"`
	Sound string                 `json:"sound,omitempty"`
}

// PushNotificationResponse represents a push notification response
type PushNotificationResponse struct {
	MessageID   string `json:"message_id"`
	Status      string `json:"status"`
	SuccessCount int   `json:"success_count"`
	FailureCount int   `json:"failure_count"`
}

// NewPushNotificationService creates a new push notification service with configurable options
// Example usage:
//   service := NewPushNotificationService(
//     WithBaseURL("https://fcm.googleapis.com/v1/projects/my-project"),
//     WithNotificationKey("your-server-key"),
//     WithTimeout(30 * time.Second),
//   )
func NewPushNotificationService(opts ...Option) *PushNotificationService {
	// Add default options for push notification service
	defaultOpts := []Option{
		WithContentType("application/json"),
		WithAccept("application/json"),
		WithUserAgent("Math-AI-Push-Client/1.0"),
		WithTimeout(30 * time.Second),
	}

	// Merge default options with provided options
	allOpts := append(defaultOpts, opts...)

	return &PushNotificationService{
		client: NewClient(allOpts...),
	}
}

// SendNotification sends a push notification via the configured provider
func (s *PushNotificationService) SendNotification(ctx context.Context, req *PushNotificationRequest) (*PushNotificationResponse, error) {
	resp, err := s.client.Post(ctx, "/messages:send", req)
	if err != nil {
		return nil, fmt.Errorf("failed to send push notification: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("push notification API error: %s", resp.String())
	}

	var notifResp PushNotificationResponse
	if err := resp.JSON(&notifResp); err != nil {
		return nil, fmt.Errorf("failed to parse push notification response: %w", err)
	}

	return &notifResp, nil
}

// SendBulkNotification sends multiple push notifications
func (s *PushNotificationService) SendBulkNotification(ctx context.Context, notifications []*PushNotificationRequest) (*PushNotificationResponse, error) {
	resp, err := s.client.Post(ctx, "/messages:send-batch", map[string]interface{}{
		"notifications": notifications,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send bulk push notifications: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("bulk push notification API error: %s", resp.String())
	}

	var notifResp PushNotificationResponse
	if err := resp.JSON(&notifResp); err != nil {
		return nil, fmt.Errorf("failed to parse bulk push notification response: %w", err)
	}

	return &notifResp, nil
}

// SendToTopic sends a push notification to a topic
func (s *PushNotificationService) SendToTopic(ctx context.Context, topic string, title, body string, data map[string]interface{}) (*PushNotificationResponse, error) {
	payload := map[string]interface{}{
		"topic": topic,
		"notification": map[string]string{
			"title": title,
			"body":  body,
		},
	}

	if data != nil {
		payload["data"] = data
	}

	resp, err := s.client.Post(ctx, "/messages:send", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to send topic notification: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("topic notification API error: %s", resp.String())
	}

	var notifResp PushNotificationResponse
	if err := resp.JSON(&notifResp); err != nil {
		return nil, fmt.Errorf("failed to parse topic notification response: %w", err)
	}

	return &notifResp, nil
}
