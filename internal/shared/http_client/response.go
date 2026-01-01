package http_client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response wraps an HTTP response with helper methods
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// IsSuccess checks if the response status code indicates success (2xx)
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError checks if the response status code indicates a client error (4xx)
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError checks if the response status code indicates a server error (5xx)
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// String returns the response body as a string
func (r *Response) String() string {
	return string(r.Body)
}

// Bytes returns the response body as bytes
func (r *Response) Bytes() []byte {
	return r.Body
}

// JSON unmarshals the response body into the provided interface
func (r *Response) JSON(v interface{}) error {
	if err := json.Unmarshal(r.Body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}
	return nil
}

// Header returns the value of the specified header
func (r *Response) Header(key string) string {
	return r.Headers.Get(key)
}

// HasHeader checks if a header exists
func (r *Response) HasHeader(key string) bool {
	_, exists := r.Headers[key]
	return exists
}

// Error returns an error if the response is not successful
func (r *Response) Error() error {
	if r.IsSuccess() {
		return nil
	}

	return &HTTPError{
		StatusCode: r.StatusCode,
		Message:    string(r.Body),
	}
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// IsHTTPError checks if an error is an HTTPError
func IsHTTPError(err error) bool {
	_, ok := err.(*HTTPError)
	return ok
}

// GetHTTPError extracts the HTTPError from an error
func GetHTTPError(err error) (*HTTPError, bool) {
	httpErr, ok := err.(*HTTPError)
	return httpErr, ok
}
