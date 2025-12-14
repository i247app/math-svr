package notification

import (
	"context"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/driven-adapter/external/http_client"
)

// PushNotificationService handles push notification-related HTTP requests
type PushNotificationService struct {
	client *http_client.Client
}

func NewPushNotificationService(opts ...http_client.Option) *PushNotificationService {
	// Add default options for push notification service
	defaultOpts := []http_client.Option{
		http_client.WithContentType("application/json"),
		http_client.WithAccept("application/json"),
		http_client.WithUserAgent("Math-AI-Push-Client/1.0"),
		http_client.WithTimeout(30 * time.Second),
	}

	// Merge default options with provided options
	allOpts := append(defaultOpts, opts...)

	return &PushNotificationService{
		client: http_client.NewClient(allOpts...),
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
