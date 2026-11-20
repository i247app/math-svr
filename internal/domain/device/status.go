package device

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (4600–4699). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	DeviceNotFound         status.StatusCode = 4601
	DeviceMissingUUID      status.StatusCode = 4602
	DeviceMissingName      status.StatusCode = 4603
	DeviceMissingUserID    status.StatusCode = 4604
	DeviceNotOwned         status.StatusCode = 4605
	DeviceAlreadyVerified  status.StatusCode = 4606
	DeviceRegistrationFail status.StatusCode = 4607
	DeviceVerificationFail status.StatusCode = 4608
	DeviceRevokeFail       status.StatusCode = 4609
	Device2FARequired      status.StatusCode = 4610
	DeviceNotTrusted       status.StatusCode = 4611
)

func init() {
	status.Register(DeviceNotFound, "Device not found", "Không tìm thấy thiết bị")
	status.Register(DeviceMissingUUID, "Device identifier is required", "Vui lòng cung cấp mã thiết bị")
	status.Register(DeviceMissingName, "Device name is required", "Vui lòng cung cấp tên thiết bị")
	status.Register(DeviceMissingUserID, "User id is required", "Mã người dùng là bắt buộc")
	status.Register(DeviceNotOwned, "Device does not belong to this user", "Thiết bị này không thuộc về người dùng")
	status.Register(DeviceAlreadyVerified, "Device is already verified", "Thiết bị đã được xác minh")
	status.Register(DeviceRegistrationFail, "Device registration failed", "Đăng ký thiết bị thất bại")
	status.Register(DeviceVerificationFail, "Device verification failed", "Xác minh thiết bị thất bại")
	status.Register(DeviceRevokeFail, "Device revoke failed", "Thu hồi thiết bị thất bại")
	status.Register(Device2FARequired, "Two-factor authentication is required for this device", "Thiết bị này yêu cầu xác thực hai yếu tố")
	status.Register(DeviceNotTrusted, "The selected device is not a trusted device", "Thiết bị được chọn chưa được tin cậy (trusted)")
}
