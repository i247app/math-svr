package notification

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
	MessageID    string `json:"message_id"`
	Status       string `json:"status"`
	SuccessCount int    `json:"success_count"`
	FailureCount int    `json:"failure_count"`
}
