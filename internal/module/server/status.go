package server

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (13000–13099). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Server lifecycle — 13000-13099
	ServerShuttingDown       status.StatusCode = 13001
	ServerReloadInvalidEnv   status.StatusCode = 13002
	ServerReloadNotSupported status.StatusCode = 13003
)

func init() {
	status.Register(ServerShuttingDown, "Server is shutting down", "Máy chủ đang tắt")
	status.Register(ServerReloadInvalidEnv, "Environment file is invalid; the server was not reloaded", "Tệp cấu hình môi trường không hợp lệ, máy chủ chưa được tải lại")
	status.Register(ServerReloadNotSupported, "Server reload is not supported on this platform", "Nền tảng này không hỗ trợ tải lại máy chủ")
}
