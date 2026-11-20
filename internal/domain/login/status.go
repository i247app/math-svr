package login

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (4200–4299). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	AuthMissingLoginName       status.StatusCode = 4201
	AuthMissingDeviceUUID      status.StatusCode = 4202
	AuthMissingIPAddress       status.StatusCode = 4203
	AuthMissingDevicePushToken status.StatusCode = 4204
	AuthInvalidToken           status.StatusCode = 4205
	AuthLoginFailed            status.StatusCode = 4206
	AuthLogoutFailed           status.StatusCode = 4207
	// AuthInvalidCredentials: /auth/login carried a password that does
	// not match the account's, or the account has no password set.
	AuthInvalidCredentials status.StatusCode = 4208
)

func init() {
	status.Register(AuthMissingLoginName, "Please enter your login name", "Vui lòng nhập tên đăng nhập")
	status.Register(AuthMissingDeviceUUID, "Device identifier is required", "Vui lòng cung cấp mã thiết bị")
	status.Register(AuthMissingIPAddress, "IP address is required", "Vui lòng cung cấp địa chỉ IP")
	status.Register(AuthMissingDevicePushToken, "Device push token is required", "Vui lòng cung cấp mã thông báo đẩy")
	status.Register(AuthInvalidToken, "Invalid or expired token", "Token không hợp lệ hoặc đã hết hạn")
	status.Register(AuthLoginFailed, "Login failed", "Đăng nhập thất bại")
	status.Register(AuthLogoutFailed, "Logout failed", "Đăng xuất thất bại")
	status.Register(AuthInvalidCredentials, "Incorrect login name or password", "Tên đăng nhập hoặc mật khẩu không đúng")
}
