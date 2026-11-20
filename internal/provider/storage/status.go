package storage

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (8000–8099). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	StorageConfigInvalid status.StatusCode = 8001
)

func init() {
	status.Register(StorageConfigInvalid, "File storage is not configured", "Chưa cấu hình dịch vụ lưu trữ tệp")
}
