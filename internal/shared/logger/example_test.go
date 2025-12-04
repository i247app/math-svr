package logger_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/logger"
)

// ExampleBasicUsage demonstrates basic logger usage without request context
func ExampleBasicUsage() {
	// Using package-level functions (will show [anon] [anon] for token and userid)
	logger.Info("This is an info message")
	logger.Infof("This is a formatted info message with value: %d", 42)
	logger.Error("This is an error message")
	logger.Warn("This is a warning message")
	logger.Debug("This is a debug message")
}

// ExampleRequestScopedUsage demonstrates request-scoped logger with session
func ExampleRequestScopedUsage() {
	// Create a mock request
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)

	// Create a mock session and add it to request context
	sess := session.NewSession()
	sess.Put("key", "abcdef123456789")  // Token
	sess.Put("uid", int64(12345))       // User ID

	// Add session to request context
	ctx := context.WithValue(req.Context(), session.SessionContextKey, sess)
	req = req.WithContext(ctx)

	// Create request-scoped logger
	log := logger.NewRequestScopedLogger(req, "")
	defer log.Close()

	// Use the logger (will show [456789] [12345] - last 6 chars of token and userid)
	log.Info("User logged in successfully")
	log.Infof("User %d accessed resource %s", 12345, "/api/test")
	log.Error("Failed to process request")
	log.Warn("Rate limit approaching")
	log.Debug("Request details processed")
}

// ExampleLoggerWithFileOutput demonstrates logging to a file
func ExampleLoggerWithFileOutput() {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)

	// Create logger that writes to a file
	log := logger.NewRequestScopedLogger(req, "/tmp/app.log")
	defer log.Close()

	log.Info("This message will be written to /tmp/app.log")
}

// TestLoggerFormat verifies the log format
func TestLoggerFormat(t *testing.T) {
	// This test demonstrates the expected format
	// Expected format: timestamp filename:line [token] [userid] LEVEL: message
	// Example: 2025/12/04 04:18:38.151018 login.go:95 [1yuhg1] [123] INFO: User login successful

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	sess := session.NewSession()
	sess.Put("key", "xyz1yuhg1")
	sess.Put("uid", int64(123))

	ctx := context.WithValue(req.Context(), session.SessionContextKey, sess)
	req = req.WithContext(ctx)

	log := logger.NewRequestScopedLogger(req, "")
	defer log.Close()

	// This will output:
	// 2025/12/05 XX:XX:XX.XXXXXX example_test.go:XX [1yuhg1] [123] INFO: User login successful
	log.Info("User login successful")
}

// TestLoggerWithoutSession verifies logger works without session
func TestLoggerWithoutSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	log := logger.NewRequestScopedLogger(req, "")
	defer log.Close()

	// This will output:
	// 2025/12/05 XX:XX:XX.XXXXXX example_test.go:XX [anon] [anon] INFO: Request without session
	log.Info("Request without session")
}

// TestPackageLevelLogger verifies package-level logger functions
func TestPackageLevelLogger(t *testing.T) {
	// This will output:
	// 2025/12/05 XX:XX:XX.XXXXXX example_test.go:XX [anon] [anon] INFO: Package level log
	logger.Info("Package level log")
	logger.Errorf("Error occurred: %v", "something went wrong")
}
