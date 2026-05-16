package status

func GetVNMessage(statusCode StatusCode) StatusMessage {
	switch statusCode {
	case SUCCESS:
		return "Thành công"
	case BAD_REQUEST:
		return "Yêu cầu không hợp lệ"
	case UNAUTHORIZED:
		return "Không có quyền truy cập"
	case FORBIDDEN:
		return "Bị từ chối"
	case NOT_FOUND:
		return "Không tìm thấy"
	case CONFLICT:
		return "Xung đột"
	case INTERNAL_SERVER_ERROR:
		return "Lỗi máy chủ nội bộ"

	// User
	case USER_MISSING_EMAIL:
		return "Vui lòng nhập email"
	case USER_MISSING_PHONE:
		return "Vui lòng nhập số điện thoại"
	case USER_MISSING_PASSWORD:
		return "Vui lòng nhập mật khẩu"
	case USER_INVALID_EMAIL:
		return "Email không hợp lệ"
	case USER_INVALID_PASSWORD:
		return "Mật khẩu không hợp lệ"
	case USER_EMAIL_ALREADY_EXISTS:
		return "Email đã tồn tại"
	case USER_PHONE_ALREADY_EXISTS:
		return "Số điện thoại đã tồn tại"
	case USER_NOT_FOUND:
		return "Không tìm thấy người dùng"

	// Notification
	case NOTIFICATION_MISSING_UID:
		return "Vui lòng cung cấp mã người dùng"
	case NOTIFICATION_MISSING_TITLE:
		return "Vui lòng nhập tiêu đề thông báo"
	case NOTIFICATION_MISSING_SHORT_TEXT:
		return "Vui lòng nhập nội dung ngắn"
	case NOTIFICATION_INVALID_ICON:
		return "Biểu tượng thông báo không hợp lệ"
	case NOTIFICATION_INVALID_PRIORITY:
		return "Mức ưu tiên không hợp lệ"
	case NOTIFICATION_NOT_FOUND:
		return "Không tìm thấy thông báo"

	// Messaging
	case MESSAGING_PUBLISH_FAILED:
		return "Gửi sự kiện thất bại"
	case MESSAGING_INVALID_ENVELOPE:
		return "Gói sự kiện không hợp lệ"
	case MESSAGING_CONFIG_INVALID:
		return "Cấu hình messaging không hợp lệ"
	case MESSAGING_CONNECT_FAILED:
		return "Kết nối broker thất bại"

	// Cache
	case CACHE_CONNECT_FAILED:
		return "Kết nối cache thất bại"
	case CACHE_CONFIG_INVALID:
		return "Cấu hình cache không hợp lệ"
	case CACHE_OP_FAILED:
		return "Thao tác cache thất bại"
	case CACHE_SERIALIZE_FAILED:
		return "Không thể mã hoá giá trị cache"

	// SMS
	case SMS_CONNECT_FAILED:
		return "Kết nối nhà cung cấp SMS thất bại"
	case SMS_CONFIG_INVALID:
		return "Cấu hình SMS không hợp lệ"
	case SMS_OP_FAILED:
		return "Thao tác SMS thất bại"
	case SMS_SERIALIZE_FAILED:
		return "Không thể tuần tự hoá payload SMS"
	default:
		return ""
	}
}
