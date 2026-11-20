package user

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (4000–4099). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	UserMissingEmail           status.StatusCode = 4001
	UserMissingPhone           status.StatusCode = 4002
	UserMissingPassword        status.StatusCode = 4003
	UserInvalidEmail           status.StatusCode = 4004
	UserInvalidPassword        status.StatusCode = 4005
	UserNotFound               status.StatusCode = 4006
	UserEmailAlreadyExists     status.StatusCode = 4007
	UserPhoneAlreadyExists     status.StatusCode = 4008
	UserAlreadyExists          status.StatusCode = 4009
	UserUsernameAlreadyExists  status.StatusCode = 4010
	UserMissingName            status.StatusCode = 4011
	UserAvatarInvalidFile      status.StatusCode = 4012
	UserAvatarUploadFailed     status.StatusCode = 4013
	UserAvatarConflict         status.StatusCode = 4014
	UserAvatarInvalidReference status.StatusCode = 4015
	UserInvalidRole            status.StatusCode = 4016
	UserEmailNotVerified       status.StatusCode = 4017
	// UserMissingDeviceUUID is the guest path's only hard requirement:
	// a guest is recognised by their device, so without one there is
	// nothing to open — or reopen — an account against.
	UserMissingDeviceUUID status.StatusCode = 4018
	// UserMissingIdentifier: /users/create needs at least one login key —
	// a phone or an email. Either alone is enough.
	UserMissingIdentifier status.StatusCode = 4019
)

func init() {
	status.Register(UserMissingEmail, "Please enter your email", "Vui lòng nhập email")
	status.Register(UserMissingPhone, "Please enter your phone", "Vui lòng nhập số điện thoại")
	status.Register(UserMissingPassword, "Please enter your password", "Vui lòng nhập mật khẩu")
	status.Register(UserInvalidEmail, "User invalid email", "Email không hợp lệ")
	status.Register(UserInvalidPassword, "Password must be between 8 and 72 characters", "Mật khẩu phải có từ 8 đến 72 ký tự")
	status.Register(UserNotFound, "User not found", "Không tìm thấy người dùng")
	status.Register(UserEmailAlreadyExists, "Email already exists", "Email đã tồn tại")
	status.Register(UserPhoneAlreadyExists, "Phone already exists", "Số điện thoại đã tồn tại")
	status.Register(UserAlreadyExists, "User already exists", "Người dùng đã tồn tại")
	status.Register(UserUsernameAlreadyExists, "Username already exists", "Username đã tồn tại")
	status.Register(UserMissingName, "Please enter your name", "Vui lòng nhập tên của bạn")
	status.Register(UserAvatarInvalidFile, "Avatar file is invalid", "Tệp ảnh đại diện không hợp lệ")
	status.Register(UserAvatarUploadFailed, "Avatar upload failed", "Tải ảnh đại diện thất bại")
	status.Register(UserAvatarConflict, "Provide either an avatar file or an avatar reference, not both", "Chỉ chọn một trong hai: tệp ảnh hoặc đường dẫn ảnh")
	status.Register(UserAvatarInvalidReference, "Avatar reference is invalid", "Đường dẫn ảnh đại diện không hợp lệ")
	status.Register(UserInvalidRole, "Role is invalid, Supported roles: {roles}", "Vai trò không hợp lệ, các vai trò được hỗ trợ: {roles}")
	status.Register(UserEmailNotVerified, "Email is not verified. Please verify your email via OTP before signing up with it", "Email chưa được xác thực. Vui lòng xác thực email bằng OTP trước khi đăng ký với email này")
	status.Register(UserMissingDeviceUUID, "Device information is missing", "Thiếu thông tin thiết bị")
	status.Register(UserMissingIdentifier, "Please enter your phone or email", "Vui lòng nhập số điện thoại hoặc email")
}
