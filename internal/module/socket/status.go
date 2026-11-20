package socket

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (13300–13399). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Socket / WebSocket — 13300-13399
	SocketUnauthorized       status.StatusCode = 13300
	SocketMissingTopic       status.StatusCode = 13301
	SocketTopicForbidden     status.StatusCode = 13302
	SocketInvalidMessage     status.StatusCode = 13303
	SocketTooManyConnections status.StatusCode = 13304
	SocketDisabled           status.StatusCode = 13305
)

func init() {
	status.Register(SocketUnauthorized, "Unauthorized WebSocket connection", "Kết nối WebSocket không được xác thực")
	status.Register(SocketMissingTopic, "topic is required", "Thiếu topic")
	status.Register(SocketTopicForbidden, "Not allowed to subscribe to this topic", "Không được phép đăng ký topic này")
	status.Register(SocketInvalidMessage, "Invalid WebSocket message", "Tin nhắn WebSocket không hợp lệ")
	status.Register(SocketTooManyConnections, "Too many active connections", "Quá nhiều kết nối đang hoạt động")
	status.Register(SocketDisabled, "Realtime channel is disabled", "Kênh realtime đang bị tắt")
}
