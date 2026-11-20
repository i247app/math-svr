package home

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (13100–13199). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Home layout — 13100-13199
	HomeMissingProfileID status.StatusCode = 13101
	HomeProfileNotFound  status.StatusCode = 13102
	HomeProfileNotOwned  status.StatusCode = 13103
	HomeUnsupportedRole  status.StatusCode = 13104
)

func init() {
	status.Register(HomeMissingProfileID, "profile_id is required", "Thiếu profile_id")
	status.Register(HomeProfileNotFound, "Profile not found", "Không tìm thấy hồ sơ")
	status.Register(HomeProfileNotOwned, "This profile does not belong to the current user", "Hồ sơ này không thuộc về người dùng hiện tại")
	status.Register(HomeUnsupportedRole, "Unsupported profile role for the home layout", "Vai trò hồ sơ không được hỗ trợ cho trang chủ")
}
