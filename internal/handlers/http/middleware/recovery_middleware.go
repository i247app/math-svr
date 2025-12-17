package middleware

import (
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/recovery"
)

// RecoveryMiddleware recovers from panics and logs them nicely
func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				// recover() MUST be called directly in the deferred function
				if rec := recover(); rec != nil {
					// Build panic info from the recovered value
					panicInfo := recovery.BuildPanicInfo(rec, 3)

					// Log the panic with nice formatting
					logPanic(r, panicInfo)

					// Send error response to client
					respondWithPanicError(w, r, panicInfo)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// logPanic logs the panic information with nice formatting
func logPanic(r *http.Request, panicInfo *recovery.PanicInfo) {
	// Print to console with nice formatting
	fmt.Print(panicInfo.Format())

	// Also log to application logger if available
	if log := logger.GetLogger(r.Context()); log != nil {
		log.Error("PANIC RECOVERED: ", panicInfo.FormatShort())
		log.Error("Full stack trace:")
		for i, frame := range panicInfo.StackTrace {
			log.Errorf("  [%d] %s at %s:%d", i+1, frame.Function, frame.File, frame.Line)
		}
	}
}

// respondWithPanicError sends an appropriate error response to the client
func respondWithPanicError(w http.ResponseWriter, r *http.Request, panicInfo *recovery.PanicInfo) {
	// Set appropriate status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	// Create error message based on panic type
	var errorMessage string
	if panicInfo.IsNilPointerError() {
		errorMessage = "Internal server error: nil pointer dereference"
	} else {
		errorMessage = "Internal server error: unexpected panic"
	}

	// Send JSON response
	response := fmt.Sprintf(`{
  "success": false,
  "error": "%s",
  "message": "The server encountered an unexpected error. Please try again or contact support.",
  "panic_type": "%T"
}`, errorMessage, panicInfo.Value)

	w.Write([]byte(response))
}

// RecoveryWithCustomHandler creates a recovery middleware with custom panic handler
func RecoveryWithCustomHandler(handler func(w http.ResponseWriter, r *http.Request, panicInfo *recovery.PanicInfo)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					panicInfo := recovery.BuildPanicInfo(rec, 3)

					// Log the panic
					logPanic(r, panicInfo)

					// Call custom handler
					handler(w, r, panicInfo)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
