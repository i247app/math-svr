package middleware

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (14200–14299). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Rate limit (bootstrap/middleware/rate_limit_middleware.go) — 14200-14299
	RateLimitExceeded status.StatusCode = 14200
)

func init() {
	status.Register(RateLimitExceeded, "Too many requests, please try again later", "Bạn đã yêu cầu vượt mức cho phép, vui lòng thử lại sau")
}
