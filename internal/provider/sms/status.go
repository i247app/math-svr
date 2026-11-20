package sms

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (9000–9099). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	SMSConnectFailed   status.StatusCode = 9001
	SMSConfigInvalid   status.StatusCode = 9002
	SMSOpFailed        status.StatusCode = 9003
	SMSSerializeFailed status.StatusCode = 9004
)

func init() {
	status.Register(SMSConnectFailed, "SMS provider connection failed", "Kết nối nhà cung cấp SMS thất bại")
	status.Register(SMSConfigInvalid, "SMS configuration is invalid", "Cấu hình SMS không hợp lệ")
	status.Register(SMSOpFailed, "SMS operation failed", "Thao tác SMS thất bại")
	status.Register(SMSSerializeFailed, "Failed to serialize SMS payload", "Không thể tuần tự hoá payload SMS")
}
