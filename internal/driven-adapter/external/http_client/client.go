package http_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"math-ai.com/math-ai/internal/shared/logger"
)

// Client is the base HTTP client for making external API calls
type Client struct {
	baseURL      string
	httpClient   *http.Client
	headers      map[string]string
	queryParams  map[string]string
	interceptors []Interceptor
	retryConfig  *RetryConfig
}

// RetryConfig defines retry behavior for failed requests
type RetryConfig struct {
	MaxRetries         int
	RetryDelay         time.Duration
	RetryableHTTPCodes []int
}

// NewClient creates a new HTTP client with the given options
func NewClient(opts ...Option) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers:      make(map[string]string),
		queryParams:  make(map[string]string),
		interceptors: []Interceptor{},
	}

	// Apply all options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil, opts...)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodPost, path, body, opts...)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodPut, path, body, opts...)
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, path string, body interface{}, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodPatch, path, body, opts...)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodDelete, path, nil, opts...)
}

// Do performs an HTTP request with the given method, path, and body
func (c *Client) Do(ctx context.Context, method, path string, body interface{}, opts ...RequestOption) (*Response, error) {
	logger := logger.GetLogger(ctx)

	// Build the request
	req, err := c.buildRequest(ctx, method, path, body, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Execute request with retry logic if configured
	if c.retryConfig != nil {
		return c.doWithRetry(ctx, req)
	}

	logger.Infof("### Call External API [%v] - %v\n", req.Method, req.URL.String())

	return c.executeRequest(ctx, req)
}

// buildRequest constructs an HTTP request with all configurations applied
func (c *Client) buildRequest(ctx context.Context, method, path string, body interface{}, opts ...RequestOption) (*http.Request, error) {
	logger := logger.GetLogger(ctx)

	// Create request config with client defaults
	reqConfig := &RequestConfig{
		headers:     make(map[string]string),
		queryParams: make(map[string]string),
	}

	// Copy client-level headers and query params
	for k, v := range c.headers {
		reqConfig.headers[k] = v
	}
	for k, v := range c.queryParams {
		reqConfig.queryParams[k] = v
	}

	// Apply request-specific options
	for _, opt := range opts {
		opt(reqConfig)
	}

	// Build URL
	url := c.baseURL + path
	if len(reqConfig.queryParams) > 0 {
		url += "?" + buildQueryString(reqConfig.queryParams)
	}

	// Encode body if present
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)

		logger.Info(string(jsonBody))

		// Set Content-Type if not already set
		if _, exists := reqConfig.headers["Content-Type"]; !exists {
			reqConfig.headers["Content-Type"] = "application/json"
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range reqConfig.headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// executeRequest executes the HTTP request and returns a Response
func (c *Client) executeRequest(ctx context.Context, req *http.Request) (*Response, error) {
	// Apply interceptors (before request)
	for _, interceptor := range c.interceptors {
		if err := interceptor.Before(req); err != nil {
			return nil, fmt.Errorf("interceptor before failed: %w", err)
		}
	}

	// Execute the request
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("### Call External API Reponse, %v\n", string(bodyBytes))

	// Create Response object
	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       bodyBytes,
	}

	// Apply interceptors (after response)
	for _, interceptor := range c.interceptors {
		if err := interceptor.After(resp); err != nil {
			return nil, fmt.Errorf("interceptor after failed: %w", err)
		}
	}

	return resp, nil
}

// doWithRetry executes the request with retry logic
func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*Response, error) {
	var resp *Response
	var err error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.retryConfig.RetryDelay)
		}

		resp, err = c.executeRequest(ctx, req)
		if err == nil && !c.shouldRetry(resp.StatusCode) {
			return resp, nil
		}
	}

	return resp, err
}

// shouldRetry determines if a request should be retried based on status code
func (c *Client) shouldRetry(statusCode int) bool {
	if c.retryConfig == nil {
		return false
	}

	for _, code := range c.retryConfig.RetryableHTTPCodes {
		if statusCode == code {
			return true
		}
	}

	// Default: retry on 5xx errors
	return statusCode >= 500 && statusCode < 600
}

// buildQueryString builds a URL query string from params
func buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	var buf bytes.Buffer
	first := true
	for key, value := range params {
		if !first {
			buf.WriteString("&")
		}
		buf.WriteString(key)
		buf.WriteString("=")
		buf.WriteString(value)
		first = false
	}
	return buf.String()
}
