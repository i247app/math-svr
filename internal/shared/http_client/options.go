package http_client

import (
	"encoding/base64"
	"net/http"
	"time"
)

// Option is a function that configures the Client
type Option func(*Client)

// RequestOption is a function that configures a specific request
type RequestOption func(*RequestConfig)

// RequestConfig holds configuration for a specific request
type RequestConfig struct {
	headers     map[string]string
	queryParams map[string]string
}

// ===============================
// Client-level Options
// ===============================

// WithBaseURL sets the base URL for all requests
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithAPIKey sets the API key header for authentication
func WithAPIKey(apiKey string, headerName ...string) Option {
	return func(c *Client) {
		key := "X-API-Key"
		if len(headerName) > 0 && headerName[0] != "" {
			key = headerName[0]
		}
		c.headers[key] = apiKey
	}
}

// WithSecretKey sets a secret key header
func WithSecretKey(secretKey string, headerName ...string) Option {
	return func(c *Client) {
		key := "X-Secret-Key"
		if len(headerName) > 0 && headerName[0] != "" {
			key = headerName[0]
		}
		c.headers[key] = secretKey
	}
}

// WithBearerToken sets the Authorization header with Bearer token
func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.headers["Authorization"] = "Bearer " + token
	}
}

// WithBasicAuth sets the Authorization header with Basic auth
func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		// Note: In production, you should use http.Request.SetBasicAuth
		// This is a simplified version
		c.headers["Authorization"] = "Basic " + encodeBasicAuth(username, password)
	}
}

// WithHeader sets a custom header for all requests
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.headers[key] = value
	}
}

// WithHeaders sets multiple custom headers
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		for key, value := range headers {
			c.headers[key] = value
		}
	}
}

// WithQueryParam sets a query parameter for all requests
func WithQueryParam(key, value string) Option {
	return func(c *Client) {
		c.queryParams[key] = value
	}
}

// WithQueryParams sets multiple query parameters
func WithQueryParams(params map[string]string) Option {
	return func(c *Client) {
		for key, value := range params {
			c.queryParams[key] = value
		}
	}
}

// WithUserAgent sets the User-Agent header
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.headers["User-Agent"] = userAgent
	}
}

// WithContentType sets the Content-Type header
func WithContentType(contentType string) Option {
	return func(c *Client) {
		c.headers["Content-Type"] = contentType
	}
}

// WithAccept sets the Accept header
func WithAccept(accept string) Option {
	return func(c *Client) {
		c.headers["Accept"] = accept
	}
}

// WithInterceptor adds an interceptor to the client
func WithInterceptor(interceptor Interceptor) Option {
	return func(c *Client) {
		c.interceptors = append(c.interceptors, interceptor)
	}
}

// WithRetry configures retry behavior
func WithRetry(maxRetries int, retryDelay time.Duration, retryableHTTPCodes ...int) Option {
	return func(c *Client) {
		c.retryConfig = &RetryConfig{
			MaxRetries:         maxRetries,
			RetryDelay:         retryDelay,
			RetryableHTTPCodes: retryableHTTPCodes,
		}
	}
}

// ===============================
// Request-level Options
// ===============================

// WithRequestHeader sets a header for a specific request
func WithRequestHeader(key, value string) RequestOption {
	return func(rc *RequestConfig) {
		rc.headers[key] = value
	}
}

// WithRequestHeaders sets multiple headers for a specific request
func WithRequestHeaders(headers map[string]string) RequestOption {
	return func(rc *RequestConfig) {
		for key, value := range headers {
			rc.headers[key] = value
		}
	}
}

// WithRequestQueryParam sets a query parameter for a specific request
func WithRequestQueryParam(key, value string) RequestOption {
	return func(rc *RequestConfig) {
		rc.queryParams[key] = value
	}
}

// WithRequestQueryParams sets multiple query parameters for a specific request
func WithRequestQueryParams(params map[string]string) RequestOption {
	return func(rc *RequestConfig) {
		for key, value := range params {
			rc.queryParams[key] = value
		}
	}
}

// ===============================
// Service-specific Options
// ===============================

// WithSMSKey sets SMS service API key (convenience wrapper)
func WithSMSKey(apiKey string) Option {
	return WithAPIKey(apiKey, "X-SMS-API-Key")
}

// WithEmailKey sets Email service API key (convenience wrapper)
func WithEmailKey(apiKey string) Option {
	return WithAPIKey(apiKey, "X-Email-API-Key")
}

// WithNotificationKey sets Notification service API key (convenience wrapper)
func WithNotificationKey(apiKey string) Option {
	return WithAPIKey(apiKey, "X-Notification-API-Key")
}

// WithTwilioAuth sets Twilio-specific authentication
func WithTwilioAuth(accountSID, authToken string) Option {
	return WithBasicAuth(accountSID, authToken)
}

// WithSendGridAuth sets SendGrid-specific authentication
func WithSendGridAuth(apiKey string) Option {
	return WithBearerToken(apiKey)
}

// WithFirebaseAuth sets Firebase-specific authentication
func WithFirebaseAuth(serverKey string) Option {
	return WithHeader("Authorization", "key="+serverKey)
}

// encodeBasicAuth encodes username and password for Basic auth
func encodeBasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
