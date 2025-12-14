# HTTP Client Package

A flexible, extensible HTTP client for making external API calls to services like email, SMS, push notifications, and more.

## Architecture Overview

This package implements a **Functional Options Pattern** for configurable HTTP clients with support for:

- Multiple authentication methods (API Key, Bearer Token, Basic Auth, Custom)
- Request/Response interceptors (logging, timing, error handling, circuit breaker, rate limiting)
- Retry logic with configurable delays and retryable HTTP codes
- Service-specific adapters (Email, SMS, Push Notifications, SMTP)
- Type-safe request/response handling

## Package Structure

```
http_client/
├── client.go           # Base HTTP client with core functionality
├── options.go          # Functional options for client configuration
├── response.go         # Response wrapper with helper methods
├── interceptor.go      # Interceptor system for middleware
├── email_service.go    # Email service adapter
├── sms_service.go      # SMS service adapter
├── push_notification_service.go  # Push notification service adapter
├── smtp_service.go     # SMTP service adapter
├── EXAMPLES.md         # Comprehensive usage examples
└── README.md           # This file
```

## Quick Start

### Basic Usage

```go
import http_client "math-ai.com/math-ai/internal/driven-adapter/external/http_client"

// Create a client
client := http_client.NewClient(
    http_client.WithBaseURL("https://api.example.com"),
    http_client.WithAPIKey("your-api-key"),
    http_client.WithTimeout(30 * time.Second),
)

// Make a request
resp, err := client.Get(ctx, "/endpoint")
```

### Using a Service

```go
// Email service
emailService := http_client.NewEmailService(
    http_client.WithBaseURL("https://api.sendgrid.com/v3"),
    http_client.WithSendGridAuth("your-api-key"),
)

emailResp, err := emailService.SendEmail(ctx, &http_client.EmailRequest{
    To:      []string{"user@example.com"},
    From:    "noreply@math-ai.com",
    Subject: "Welcome!",
    Body:    "Thank you for signing up",
})
```

## Core Components

### 1. Client (`client.go`)

The base HTTP client that handles all HTTP operations. Supports:
- GET, POST, PUT, PATCH, DELETE methods
- Automatic retry with configurable backoff
- Interceptor execution
- Request/response handling

### 2. Options (`options.go`)

Functional options for configuring clients:

#### Client-Level Options
- `WithBaseURL(url)` - Set base URL
- `WithTimeout(duration)` - Set request timeout
- `WithAPIKey(key, headerName?)` - Set API key authentication
- `WithBearerToken(token)` - Set Bearer token authentication
- `WithBasicAuth(user, pass)` - Set Basic authentication
- `WithHeader(key, value)` - Add custom header
- `WithHeaders(map)` - Add multiple headers
- `WithQueryParam(key, value)` - Add query parameter
- `WithRetry(maxRetries, delay, codes...)` - Configure retry behavior
- `WithInterceptor(interceptor)` - Add interceptor

#### Request-Level Options
- `WithRequestHeader(key, value)` - Override header for specific request
- `WithRequestQueryParam(key, value)` - Add query param to specific request

#### Service-Specific Options
- `WithSMSKey(key)` - SMS service API key
- `WithEmailKey(key)` - Email service API key
- `WithNotificationKey(key)` - Notification service API key
- `WithTwilioAuth(sid, token)` - Twilio authentication
- `WithSendGridAuth(key)` - SendGrid authentication
- `WithFirebaseAuth(key)` - Firebase authentication

### 3. Response (`response.go`)

Response wrapper with helper methods:
- `IsSuccess()` - Check if 2xx status
- `IsClientError()` - Check if 4xx status
- `IsServerError()` - Check if 5xx status
- `String()` - Get body as string
- `Bytes()` - Get body as bytes
- `JSON(v)` - Unmarshal JSON into struct
- `Error()` - Get error if not successful

### 4. Interceptors (`interceptor.go`)

Middleware for cross-cutting concerns:

- **LoggingInterceptor** - Log request/response details
- **TimingInterceptor** - Measure request duration
- **ErrorHandlingInterceptor** - Custom error handling
- **RateLimitInterceptor** - Rate limiting
- **CircuitBreakerInterceptor** - Circuit breaker pattern
- **CustomHeaderInterceptor** - Add headers dynamically
- **ValidationInterceptor** - Validate responses

### 5. Services

Pre-built service adapters:

- **EmailService** - Email sending (SendGrid, Mailgun, etc.)
- **SmsService** - SMS sending (Twilio, etc.)
- **PushNotificationService** - Push notifications (FCM, etc.)
- **SMTPService** - SMTP email with attachments

## Design Patterns

### Functional Options Pattern

Clean, extensible configuration:

```go
service := NewEmailService(
    WithBaseURL("https://api.provider.com"),
    WithAPIKey("key"),
    WithTimeout(30 * time.Second),
    WithRetry(3, 2*time.Second),
    WithInterceptor(NewLoggingInterceptor(true)),
)
```

### Interceptor Pattern

Chain of responsibility for request/response processing:

```go
type Interceptor interface {
    Before(req *http.Request) error
    After(resp *Response) error
}
```

### Adapter Pattern

Service-specific adapters wrap the base client:

```go
type EmailService struct {
    client *Client
}

func (s *EmailService) SendEmail(ctx context.Context, req *EmailRequest) (*EmailResponse, error) {
    resp, err := s.client.Post(ctx, "/send", req)
    // ... handle response
}
```

## Key Features

### 1. Flexible Authentication

Multiple authentication strategies:

```go
// API Key
WithAPIKey("key", "X-API-Key")

// Bearer Token
WithBearerToken("jwt-token")

// Basic Auth
WithBasicAuth("user", "pass")

// Custom Header
WithHeader("X-Custom-Auth", "value")
```

### 2. Retry Logic

Automatic retry with exponential backoff:

```go
WithRetry(
    3,                    // max retries
    2 * time.Second,     // delay between retries
    500, 502, 503,       // retryable HTTP codes
)
```

### 3. Circuit Breaker

Prevent cascading failures:

```go
WithInterceptor(
    NewCircuitBreakerInterceptor(
        5,                // failure threshold
        30 * time.Second, // reset timeout
    ),
)
```

### 4. Rate Limiting

Control request rate:

```go
WithInterceptor(
    NewRateLimitInterceptor(10), // 10 requests per second
)
```

### 5. Request/Response Logging

Debug API calls:

```go
WithInterceptor(
    NewLoggingInterceptor(true), // log bodies
)
```

## Adding a New Service

To add a new service (e.g., Payment):

1. Create `payment_service.go`:

```go
type PaymentService struct {
    client *Client
}

func NewPaymentService(opts ...Option) *PaymentService {
    defaultOpts := []Option{
        WithContentType("application/json"),
        WithTimeout(30 * time.Second),
    }
    return &PaymentService{
        client: NewClient(append(defaultOpts, opts...)...),
    }
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    resp, err := s.client.Post(ctx, "/payments", req)
    // ... handle response
}
```

2. Define request/response types:

```go
type PaymentRequest struct {
    Amount   int64  `json:"amount"`
    Currency string `json:"currency"`
}

type PaymentResponse struct {
    PaymentID string `json:"payment_id"`
    Status    string `json:"status"`
}
```

3. Use it:

```go
paymentService := NewPaymentService(
    WithBaseURL("https://api.stripe.com"),
    WithBearerToken("sk_test_..."),
)

resp, err := paymentService.CreatePayment(ctx, &PaymentRequest{
    Amount:   1000,
    Currency: "usd",
})
```

## Testing

### Mocking the Client

```go
// Create a mock server
mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}))
defer mockServer.Close()

// Use mock server URL
client := NewClient(
    WithBaseURL(mockServer.URL),
)
```

### Testing with Interceptors

```go
func TestWithLogging(t *testing.T) {
    client := NewClient(
        WithBaseURL("https://api.example.com"),
        WithInterceptor(NewLoggingInterceptor(true)),
    )

    // Test requests - all will be logged
}
```

## Best Practices

1. **Use Service Adapters**: Don't use the base client directly; create service-specific adapters
2. **Set Timeouts**: Always configure appropriate timeouts
3. **Add Logging**: Use LoggingInterceptor in development
4. **Handle Errors**: Check response status and handle errors appropriately
5. **Use Context**: Always pass context for cancellation support
6. **Configure Retry**: For unreliable APIs, configure retry logic
7. **Rate Limit**: Respect API rate limits using RateLimitInterceptor
8. **Circuit Breaker**: Use for external services that may fail

## Examples

See [EXAMPLES.md](./EXAMPLES.md) for comprehensive usage examples including:
- Email service (SendGrid, Mailgun)
- SMS service (Twilio)
- Push notifications (FCM)
- SMTP service
- Custom interceptors
- Adding new services

## Performance Considerations

- **Connection Pooling**: The underlying `http.Client` handles connection pooling
- **Timeouts**: Set appropriate timeouts to prevent hanging requests
- **Rate Limiting**: Use `RateLimitInterceptor` to avoid overwhelming external APIs
- **Circuit Breaker**: Prevent cascading failures with `CircuitBreakerInterceptor`
- **Retry Logic**: Configure smart retry with exponential backoff

## Error Handling

```go
resp, err := client.Get(ctx, "/endpoint")
if err != nil {
    // Network error or other failure
    log.Fatal(err)
}

if !resp.IsSuccess() {
    // HTTP error (4xx, 5xx)
    if httpErr, ok := http_client.GetHTTPError(resp.Error()); ok {
        log.Printf("HTTP Error %d: %s", httpErr.StatusCode, httpErr.Message)
    }
}
```

## Contributing

When adding new features:
1. Follow the existing patterns (Functional Options, Interceptors)
2. Add comprehensive tests
3. Update documentation and examples
4. Ensure backward compatibility

## License

Internal use for Math-AI project.
