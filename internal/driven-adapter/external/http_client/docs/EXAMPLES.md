# HTTP Client Examples

This document demonstrates how to use the flexible HTTP client system for external API integrations.

## Table of Contents

1. [Basic Usage](#basic-usage)
2. [Email Service Examples](#email-service-examples)
3. [SMS Service Examples](#sms-service-examples)
4. [Push Notification Service Examples](#push-notification-service-examples)
5. [SMTP Service Examples](#smtp-service-examples)
6. [Advanced Options](#advanced-options)
7. [Interceptors](#interceptors)
8. [Adding New Services](#adding-new-services)

---

## Basic Usage

### Creating a Basic HTTP Client

```go
import (
    "context"
    "time"
    http_client "math-ai.com/math-ai/internal/driven-adapter/external/http_client"
)

// Simple client with API key
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithAPIKey("your-api-key"),
    http_client.WithTimeout(30 * time.Second),
)

// Make a GET request
resp, err := client.Get(context.Background(), "/users/123")
if err != nil {
    log.Fatal(err)
}

// Parse JSON response
var user User
if err := resp.JSON(&user); err != nil {
    log.Fatal(err)
}
```

---

## Email Service Examples

### Example 1: SendGrid Integration

```go
// Create email service with SendGrid
emailService := http_client.NewEmailService(
    http_client.WithBaseURL("https://api.sendgrid.com/v3"),
    http_client.WithSendGridAuth("your-sendgrid-api-key"),
)

// Send an email
req := &http_client.EmailRequest{
    To:      []string{"user@example.com"},
    From:    "noreply@math-ai.com",
    Subject: "Welcome to Math AI",
    Body:    "Thank you for signing up!",
    HTML:    "<h1>Welcome!</h1><p>Thank you for signing up!</p>",
}

resp, err := emailService.SendEmail(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Email sent! Message ID: %s\n", resp.MessageID)
```

### Example 2: Custom Email Provider

```go
emailService := http_client.NewEmailService(
    http_client.WithBaseURL("https://api.custommail.com"),
    http_client.WithEmailKey("your-custom-key"),
    http_client.WithTimeout(60 * time.Second),
    http_client.WithHeader("X-Custom-Header", "custom-value"),
)

// Send bulk emails
emails := []*http_client.EmailRequest{
    {To: []string{"user1@example.com"}, From: "noreply@math-ai.com", Subject: "Hello 1"},
    {To: []string{"user2@example.com"}, From: "noreply@math-ai.com", Subject: "Hello 2"},
}

err := emailService.SendBulkEmail(context.Background(), emails)
```

---

## SMS Service Examples

### Example 1: Twilio Integration

```go
// Create SMS service with Twilio
smsService := http_client.NewSmsService(
    http_client.WithBaseURL("https://api.twilio.com/2010-04-01"),
    http_client.WithTwilioAuth("ACCOUNT_SID", "AUTH_TOKEN"),
)

// Send SMS
req := &http_client.SMSRequest{
    To:      "+1234567890",
    From:    "+0987654321",
    Message: "Your verification code is: 123456",
}

resp, err := smsService.SendSMS(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("SMS sent! Message ID: %s, Cost: %s\n", resp.MessageID, resp.Cost)
```

### Example 2: Custom SMS Provider

```go
smsService := http_client.NewSmsService(
    http_client.WithBaseURL("https://api.smsprovider.com"),
    http_client.WithSMSKey("your-sms-api-key"),
    http_client.WithRetry(3, 2*time.Second, 500, 502, 503),
)

// Send bulk SMS
messages := []*http_client.SMSRequest{
    {To: "+1111111111", From: "+0000000000", Message: "Message 1"},
    {To: "+2222222222", From: "+0000000000", Message: "Message 2"},
}

responses, err := smsService.SendBulkSMS(context.Background(), messages)
```

---

## Push Notification Service Examples

### Example 1: Firebase Cloud Messaging (FCM)

```go
// Create push notification service with FCM
notifService := http_client.NewPushNotificationService(
    http_client.WithBaseURL("https://fcm.googleapis.com/v1/projects/my-project"),
    http_client.WithFirebaseAuth("your-server-key"),
)

// Send notification
req := &http_client.PushNotificationRequest{
    To:    []string{"device-token-1", "device-token-2"},
    Title: "New Quiz Available!",
    Body:  "Try our new math quiz on algebra",
    Data: map[string]interface{}{
        "quiz_id": "123",
        "level":   "advanced",
    },
    Badge: 1,
    Sound: "default",
}

resp, err := notifService.SendNotification(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Notification sent! Success: %d, Failed: %d\n",
    resp.SuccessCount, resp.FailureCount)
```

### Example 2: Send to Topic

```go
resp, err := notifService.SendToTopic(
    context.Background(),
    "math-quiz-updates",
    "New Feature!",
    "Check out our new practice mode",
    map[string]interface{}{
        "feature": "practice-mode",
    },
)
```

---

## SMTP Service Examples

### Example 1: Mailgun Integration

```go
smtpService := http_client.NewSMTPService(
    http_client.WithBaseURL("https://api.mailgun.net"),
    http_client.WithAPIKey("your-mailgun-api-key"),
)

// Send email with attachment
req := &http_client.SMTPEmailRequest{
    To:       []string{"user@example.com"},
    Cc:       []string{"manager@example.com"},
    From:     "noreply@math-ai.com",
    Subject:  "Quiz Results",
    TextBody: "Your quiz results are attached.",
    HTMLBody: "<h2>Quiz Results</h2><p>Your results are attached.</p>",
    Attachments: []http_client.SMTPAttachment{
        {
            Filename:    "results.pdf",
            ContentType: "application/pdf",
            Content:     "base64-encoded-content-here",
        },
    },
}

resp, err := smtpService.SendEmail(context.Background(), req)
```

### Example 2: Template Email

```go
resp, err := smtpService.SendTemplate(
    context.Background(),
    "welcome-email-template",
    []string{"newuser@example.com"},
    map[string]interface{}{
        "username":     "John Doe",
        "login_url":    "https://math-ai.com/login",
        "support_email": "support@math-ai.com",
    },
)
```

---

## Advanced Options

### Using Multiple Authentication Methods

```go
// Bearer Token
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithBearerToken("your-jwt-token"),
)

// Basic Auth
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithBasicAuth("username", "password"),
)

// Secret Key
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithSecretKey("your-secret", "X-Secret-Key"),
)
```

### Custom Headers and Query Params

```go
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithHeaders(map[string]string{
        "X-Custom-Header": "value",
        "X-API-Version":   "v2",
    }),
    http_client.WithQueryParams(map[string]string{
        "source": "mobile-app",
        "version": "1.0.0",
    }),
)
```

### Request-Level Options

```go
// Override headers for a specific request
resp, err := client.Get(
    context.Background(),
    "/special-endpoint",
    http_client.WithRequestHeader("X-Special", "override-value"),
    http_client.WithRequestQueryParam("filter", "active"),
)
```

### Retry Configuration

```go
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithAPIKey("key"),
    // Retry 3 times with 2-second delay on 500, 502, 503 errors
    http_client.WithRetry(3, 2*time.Second, 500, 502, 503),
)
```

---

## Interceptors

### Logging Interceptor

```go
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithInterceptor(
        http_client.NewLoggingInterceptor(true), // Log request/response bodies
    ),
)
```

### Timing Interceptor

```go
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithInterceptor(
        http_client.NewTimingInterceptor(),
    ),
)
```

### Error Handling Interceptor

```go
errorHandler := func(resp *http_client.Response) error {
    if resp.StatusCode == 401 {
        return fmt.Errorf("authentication failed")
    }
    if resp.StatusCode == 429 {
        return fmt.Errorf("rate limit exceeded")
    }
    return nil
}

client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithInterceptor(
        http_client.NewErrorHandlingInterceptor(errorHandler),
    ),
)
```

### Circuit Breaker Interceptor

```go
// Open circuit after 5 failures, reset after 30 seconds
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithInterceptor(
        http_client.NewCircuitBreakerInterceptor(5, 30*time.Second),
    ),
)
```

### Rate Limiting Interceptor

```go
// Limit to 10 requests per second
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithInterceptor(
        http_client.NewRateLimitInterceptor(10),
    ),
)
```

### Multiple Interceptors

```go
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithInterceptor(http_client.NewLoggingInterceptor(false)),
    http_client.WithInterceptor(http_client.NewTimingInterceptor()),
    http_client.WithInterceptor(http_client.NewRateLimitInterceptor(10)),
    http_client.WithInterceptor(http_client.NewCircuitBreakerInterceptor(5, 30*time.Second)),
)
```

---

## Adding New Services

### Example: Adding a Payment Service

```go
package http_client

import (
    "context"
    "fmt"
    "time"
)

// PaymentService handles payment-related HTTP requests
type PaymentService struct {
    client *Client
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
    Amount      int64  `json:"amount"`
    Currency    string `json:"currency"`
    Description string `json:"description"`
    CustomerID  string `json:"customer_id"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
    PaymentID string `json:"payment_id"`
    Status    string `json:"status"`
    Amount    int64  `json:"amount"`
}

// NewPaymentService creates a new payment service
func NewPaymentService(opts ...Option) *PaymentService {
    defaultOpts := []Option{
        WithContentType("application/json"),
        WithAccept("application/json"),
        WithUserAgent("Math-AI-Payment-Client/1.0"),
        WithTimeout(45 * time.Second),
    }

    allOpts := append(defaultOpts, opts...)

    return &PaymentService{
        client: NewClient(allOpts...),
    }
}

// CreatePayment creates a new payment
func (s *PaymentService) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    resp, err := s.client.Post(ctx, "/v1/payments", req)
    if err != nil {
        return nil, fmt.Errorf("failed to create payment: %w", err)
    }

    if !resp.IsSuccess() {
        return nil, fmt.Errorf("payment API error: %s", resp.String())
    }

    var paymentResp PaymentResponse
    if err := resp.JSON(&paymentResp); err != nil {
        return nil, fmt.Errorf("failed to parse payment response: %w", err)
    }

    return &paymentResp, nil
}

// GetPayment retrieves payment details
func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
    resp, err := s.client.Get(ctx, fmt.Sprintf("/v1/payments/%s", paymentID))
    if err != nil {
        return nil, fmt.Errorf("failed to get payment: %w", err)
    }

    if !resp.IsSuccess() {
        return nil, fmt.Errorf("payment API error: %s", resp.String())
    }

    var paymentResp PaymentResponse
    if err := resp.JSON(&paymentResp); err != nil {
        return nil, fmt.Errorf("failed to parse payment response: %w", err)
    }

    return &paymentResp, nil
}
```

### Usage of New Payment Service

```go
// Create payment service with Stripe
paymentService := http_client.NewPaymentService(
    http_client.WithBaseURL("https://api.stripe.com"),
    http_client.WithBearerToken("sk_test_..."),
    http_client.WithInterceptor(http_client.NewLoggingInterceptor(true)),
)

// Create a payment
req := &http_client.PaymentRequest{
    Amount:      1000, // $10.00
    Currency:    "usd",
    Description: "Math AI Premium Subscription",
    CustomerID:  "cus_123456",
}

resp, err := paymentService.CreatePayment(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Payment created! ID: %s, Status: %s\n", resp.PaymentID, resp.Status)
```

---

## Design Patterns Used

1. **Functional Options Pattern**: Flexible configuration via `WithXxx()` functions
2. **Builder Pattern**: Request building with fluent API
3. **Interceptor/Chain of Responsibility Pattern**: Middleware for cross-cutting concerns
4. **Strategy Pattern**: Different authentication strategies
5. **Adapter Pattern**: Service-specific adapters wrapping the base client

## Benefits

- **Extensibility**: Easy to add new services without modifying core client
- **Flexibility**: Configure services differently using option functions
- **Testability**: Easy to mock and test with dependency injection
- **Maintainability**: Clear separation of concerns
- **Reusability**: Common functionality in base client, shared across all services
