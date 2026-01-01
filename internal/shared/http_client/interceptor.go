package http_client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"math-ai.com/math-ai/internal/shared/logger"
)

// Interceptor defines the interface for request/response interceptors
type Interceptor interface {
	Before(ctx context.Context, req *http.Request) error
	After(ctx context.Context, resp *Response) error
}

// ===============================
// Logging Interceptor
// ===============================

// LoggingInterceptor logs request and response details
type LoggingInterceptor struct {
	LogBody bool
}

// NewLoggingInterceptor creates a new logging interceptor
func NewLoggingInterceptor(logBody bool) *LoggingInterceptor {
	return &LoggingInterceptor{
		LogBody: logBody,
	}
}

// Before logs the request details
func (i *LoggingInterceptor) Before(ctx context.Context, req *http.Request) error {
	logger := logger.GetLogger(ctx)

	logger.Infof("[HTTP Client] %s %s", req.Method, req.URL.String())

	if i.LogBody && req.Body != nil {
		logger.Debugf("[HTTP Client] Request Headers: %v", req.Header)
	}

	return nil
}

// After logs the response details
func (i *LoggingInterceptor) After(ctx context.Context, resp *Response) error {
	logger := logger.GetLogger(ctx)

	logger.Infof("[HTTP Client] Response Status: %d", resp.StatusCode)

	if i.LogBody {
		logger.Debugf("[HTTP Client] Response Body: %s", string(resp.Body))
	}

	return nil
}

// ===============================
// Timing Interceptor
// ===============================

// TimingInterceptor measures request execution time
type TimingInterceptor struct {
	startTime time.Time
}

// NewTimingInterceptor creates a new timing interceptor
func NewTimingInterceptor() *TimingInterceptor {
	return &TimingInterceptor{}
}

// Before records the start time
func (i *TimingInterceptor) Before(ctx context.Context, req *http.Request) error {
	i.startTime = time.Now()
	return nil
}

// After logs the elapsed time
func (i *TimingInterceptor) After(ctx context.Context, resp *Response) error {
	logger := logger.GetLogger(ctx)

	elapsed := time.Since(i.startTime)
	logger.Infof("[HTTP Client] Request completed in %s", elapsed)
	return nil
}

// ===============================
// Error Handling Interceptor
// ===============================

// ErrorHandlingInterceptor handles HTTP errors
type ErrorHandlingInterceptor struct {
	OnError func(resp *Response) error
}

// NewErrorHandlingInterceptor creates a new error handling interceptor
func NewErrorHandlingInterceptor(onError func(resp *Response) error) *ErrorHandlingInterceptor {
	return &ErrorHandlingInterceptor{
		OnError: onError,
	}
}

// Before does nothing
func (i *ErrorHandlingInterceptor) Before(ctx context.Context, req *http.Request) error {
	return nil
}

// After checks for errors and invokes the error handler
func (i *ErrorHandlingInterceptor) After(ctx context.Context, resp *Response) error {
	if !resp.IsSuccess() && i.OnError != nil {
		return i.OnError(resp)
	}
	return nil
}

// ===============================
// Rate Limiting Interceptor
// ===============================

// RateLimitInterceptor implements rate limiting
type RateLimitInterceptor struct {
	requestsPerSecond int
	lastRequest       time.Time
}

// NewRateLimitInterceptor creates a new rate limiting interceptor
func NewRateLimitInterceptor(requestsPerSecond int) *RateLimitInterceptor {
	return &RateLimitInterceptor{
		requestsPerSecond: requestsPerSecond,
		lastRequest:       time.Now(),
	}
}

// Before enforces rate limiting
func (i *RateLimitInterceptor) Before(ctx context.Context, req *http.Request) error {
	minInterval := time.Second / time.Duration(i.requestsPerSecond)
	elapsed := time.Since(i.lastRequest)

	if elapsed < minInterval {
		time.Sleep(minInterval - elapsed)
	}

	i.lastRequest = time.Now()
	return nil
}

// After does nothing
func (i *RateLimitInterceptor) After(ctx context.Context, resp *Response) error {
	return nil
}

// ===============================
// Custom Header Interceptor
// ===============================

// CustomHeaderInterceptor adds custom headers to requests
type CustomHeaderInterceptor struct {
	Headers map[string]string
}

// NewCustomHeaderInterceptor creates a new custom header interceptor
func NewCustomHeaderInterceptor(headers map[string]string) *CustomHeaderInterceptor {
	return &CustomHeaderInterceptor{
		Headers: headers,
	}
}

// Before adds custom headers to the request
func (i *CustomHeaderInterceptor) Before(ctx context.Context, req *http.Request) error {
	for key, value := range i.Headers {
		req.Header.Set(key, value)
	}
	return nil
}

// After does nothing
func (i *CustomHeaderInterceptor) After(ctx context.Context, resp *Response) error {
	return nil
}

// ===============================
// Validation Interceptor
// ===============================

// ValidationInterceptor validates responses
type ValidationInterceptor struct {
	ValidateFunc func(resp *Response) error
}

// NewValidationInterceptor creates a new validation interceptor
func NewValidationInterceptor(validateFunc func(resp *Response) error) *ValidationInterceptor {
	return &ValidationInterceptor{
		ValidateFunc: validateFunc,
	}
}

// Before does nothing
func (i *ValidationInterceptor) Before(ctx context.Context, req *http.Request) error {
	return nil
}

// After validates the response
func (i *ValidationInterceptor) After(ctx context.Context, resp *Response) error {
	if i.ValidateFunc != nil {
		return i.ValidateFunc(resp)
	}
	return nil
}

// ===============================
// Circuit Breaker Interceptor
// ===============================

// CircuitBreakerInterceptor implements circuit breaker pattern
type CircuitBreakerInterceptor struct {
	failureThreshold int
	resetTimeout     time.Duration
	failureCount     int
	lastFailureTime  time.Time
	state            string // "closed", "open", "half-open"
}

// NewCircuitBreakerInterceptor creates a new circuit breaker interceptor
func NewCircuitBreakerInterceptor(failureThreshold int, resetTimeout time.Duration) *CircuitBreakerInterceptor {
	return &CircuitBreakerInterceptor{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            "closed",
	}
}

// Before checks if the circuit is open
func (i *CircuitBreakerInterceptor) Before(ctx context.Context, req *http.Request) error {
	if i.state == "open" {
		if time.Since(i.lastFailureTime) > i.resetTimeout {
			i.state = "half-open"
			i.failureCount = 0
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}
	return nil
}

// After updates the circuit breaker state based on response
func (i *CircuitBreakerInterceptor) After(ctx context.Context, resp *Response) error {
	logger := logger.GetLogger(ctx)
	if resp.IsServerError() {
		i.failureCount++
		i.lastFailureTime = time.Now()

		if i.failureCount >= i.failureThreshold {
			i.state = "open"
			logger.Warnf("[Circuit Breaker] Circuit opened after %d failures", i.failureCount)
		}
	} else if resp.IsSuccess() && i.state == "half-open" {
		i.state = "closed"
		i.failureCount = 0
		logger.Infof("[Circuit Breaker] Circuit closed")
	}

	return nil
}
