package otp

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (4700–4799). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	OTPNotFound                     status.StatusCode = 4701
	OTPMissingType                  status.StatusCode = 4702
	OTPMissingIdentifier            status.StatusCode = 4703
	OTPMissingCode                  status.StatusCode = 4704
	OTPInvalidType                  status.StatusCode = 4705
	OTPInvalidCode                  status.StatusCode = 4706
	OTPExpired                      status.StatusCode = 4707
	OTPAlreadyVerified              status.StatusCode = 4708
	OTPRevoked                      status.StatusCode = 4709
	OTPTooManyAttempts              status.StatusCode = 4710
	OTPTooFrequent                  status.StatusCode = 4711
	OTPRateLimited                  status.StatusCode = 4712
	OTPDeliveryFailed               status.StatusCode = 4713
	OTPNoDeliveryChannel            status.StatusCode = 4714
	OTPGenerationFailed             status.StatusCode = 4715
	OTPDisabled                     status.StatusCode = 4716
	OTPTargetDeviceRequiresLogin2FA status.StatusCode = 4717
)

func init() {
	status.Register(OTPNotFound, "OTP not found", "Không tìm thấy mã OTP")
	status.Register(OTPMissingType, "OTP type is required", "Vui lòng cung cấp loại OTP")
	status.Register(OTPMissingIdentifier, "Phone or email is required", "Vui lòng nhập số điện thoại hoặc email")
	status.Register(OTPMissingCode, "OTP code is required", "Vui lòng nhập mã OTP")
	status.Register(OTPInvalidType, "OTP type is invalid", "Loại OTP không hợp lệ")
	status.Register(OTPInvalidCode, "OTP code is invalid", "Mã OTP không chính xác")
	status.Register(OTPExpired, "OTP has expired", "Mã OTP đã hết hạn")
	status.Register(OTPAlreadyVerified, "OTP has already been verified", "Mã OTP đã được xác thực")
	status.Register(OTPRevoked, "OTP has been revoked", "Mã OTP đã bị thu hồi")
	status.Register(OTPTooManyAttempts, "Too many verification attempts; request a new OTP", "Bạn đã nhập sai quá nhiều lần; vui lòng yêu cầu mã OTP mới")
	status.Register(OTPTooFrequent, "Please wait before requesting another OTP", "Vui lòng đợi trước khi yêu cầu mã OTP mới")
	status.Register(OTPRateLimited, "OTP request limit reached; try again later", "Đã vượt quá số lần yêu cầu OTP; vui lòng thử lại sau")
	status.Register(OTPDeliveryFailed, "OTP delivery failed", "Gửi mã OTP thất bại")
	status.Register(OTPNoDeliveryChannel, "No delivery channel is configured for this identifier", "Không có kênh gửi phù hợp cho thông tin này")
	status.Register(OTPGenerationFailed, "Failed to generate OTP", "Không thể tạo mã OTP")
	status.Register(OTPDisabled, "OTP is disabled", "Xác thực 2 bước đã bị tắt")
	status.Register(OTPTargetDeviceRequiresLogin2FA, "target_device_id is only supported for LOGIN_2FA", "target_device_id chỉ được hỗ trợ với type LOGIN_2FA")
}
