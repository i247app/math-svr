package presence

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (13600–13699). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Presence (realtime online/offline) — 13600-13699
	PresenceNotFound     status.StatusCode = 13600
	PresenceUpdateFailed status.StatusCode = 13601
)

func init() {
	status.Register(PresenceNotFound, "Presence record not found", "Không tìm thấy bản ghi trạng thái hoạt động")
	status.Register(PresenceUpdateFailed, "Failed to update presence", "Cập nhật trạng thái hoạt động thất bại")
}
