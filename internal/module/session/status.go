package session

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (14100–14199). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Session (in-memory session store, /sessions/*) — 14100-14199
	SessionMissingToken    status.StatusCode = 14100
	SessionNotFound        status.StatusCode = 14101 // no live session is stored under that token
	SessionMissingIsSecure status.StatusCode = 14102
	SessionMissingUID      status.StatusCode = 14103 // is_secure: true on a session no user is signed into
)

func init() {
	status.Register(SessionMissingToken, "Please send the session token", "Vui lòng gửi token của phiên đăng nhập")
	status.Register(SessionNotFound, "Session not found", "Không tìm thấy phiên đăng nhập")
	status.Register(SessionMissingIsSecure, "Please send is_secure (true or false)", "Vui lòng gửi is_secure (true hoặc false)")
	status.Register(SessionMissingUID, "Only a session with a signed-in user can be marked secure", "Chỉ phiên đã có người dùng đăng nhập mới được đánh dấu bảo mật")
}
