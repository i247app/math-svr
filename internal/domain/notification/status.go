package notification

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (5000–5099). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	NotificationMissingUID       status.StatusCode = 5001
	NotificationMissingTitle     status.StatusCode = 5002
	NotificationMissingShortText status.StatusCode = 5003
	NotificationInvalidIcon      status.StatusCode = 5004
	NotificationInvalidPriority  status.StatusCode = 5005
	NotificationNotFound         status.StatusCode = 5006
	NotificationSendFailed       status.StatusCode = 5007
	NotificationNoDeviceToken    status.StatusCode = 5008
	NotificationDisabled         status.StatusCode = 5009
	NotificationNotOwned         status.StatusCode = 5010
	NotificationConfigInvalid    status.StatusCode = 5011
	NotificationConnectFailed    status.StatusCode = 5012
)

func init() {
	status.Register(NotificationMissingUID, "User id is required", "Vui lòng cung cấp mã người dùng")
	status.Register(NotificationMissingTitle, "Notification title is required", "Vui lòng nhập tiêu đề thông báo")
	status.Register(NotificationMissingShortText, "Notification short text is required", "Vui lòng nhập nội dung ngắn")
	status.Register(NotificationInvalidIcon, "Notification icon is invalid", "Biểu tượng thông báo không hợp lệ")
	status.Register(NotificationInvalidPriority, "Notification priority is invalid", "Mức ưu tiên không hợp lệ")
	status.Register(NotificationNotFound, "Notification not found", "Không tìm thấy thông báo")
	status.Register(NotificationSendFailed, "Failed to send notification", "Gửi thông báo thất bại")
	status.Register(NotificationNoDeviceToken, "No device token available for the recipient", "Người nhận không có thiết bị nào để nhận thông báo")
	status.Register(NotificationDisabled, "Notification service is disabled", "Dịch vụ thông báo đang bị tắt")
	status.Register(NotificationNotOwned, "Notification does not belong to the current user", "Thông báo không thuộc về người dùng hiện tại")
	status.Register(NotificationConfigInvalid, "Notification configuration is invalid", "Cấu hình thông báo không hợp lệ")
	status.Register(NotificationConnectFailed, "Failed to connect to the notification provider", "Không thể kết nối tới dịch vụ thông báo")
}
