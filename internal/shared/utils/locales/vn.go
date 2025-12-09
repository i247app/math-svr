package locales

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

var (
	VN LanguageType = "vn"
)

func GetMessageVNFromStatus(statusCode status.Code) string {
	args := GetArgsByStatatus(statusCode)

	switch statusCode {
	case status.OK:
		return "Thành công."
	case status.CREATED:
		return "Tài nguyên đã được tạo thành công."
	case status.FAIL:
		return "Yêu cầu thất bại."
	case status.UNAUTHORIZED:
		return "Truy cập không được phép."
	case status.NOT_FOUND:
		return "Không tìm thấy tài nguyên."
	case status.INTERNAL:
		return "Lỗi máy chủ nội bộ."

	// User status messages
	case status.USER_MISSING_ID:
		return "Thiếu ID người dùng."
	case status.USER_INVALID_PARAMS:
		return "Tham số người dùng không hợp lệ."
	case status.USER_INVALID_ID:
		return "ID người dùng không hợp lệ."
	case status.USER_NOT_FOUND:
		return "Không tìm thấy người dùng."
	case status.USER_MISSING_NAME:
		return "Thiếu tên người dùng."
	case status.USER_MISSING_EMAIL:
		return "Thiếu email người dùng."
	case status.USER_INVALID_EMAIL:
		return "Định dạng email người dùng không hợp lệ."
	case status.USER_EMAIL_ALREADY_EXISTS:
		return "Email đã tồn tại."
	case status.USER_MISSING_PHONE:
		return "Thiếu số điện thoại người dùng."
	case status.USER_PHONE_ALREADY_EXISTS:
		return "Số điện thoại đã tồn tại."
	case status.USER_INVALID_PHONE:
		return "Số điện thoại người dùng không hợp lệ."
	case status.USER_INVALID_ROLE:
		return fmt.Sprintf("Vai trò không hợp lệ. Các vai trò hợp lệ là: %v", args)
	case status.USER_INVALID_STATUS:
		return "Trạng thái người dùng không hợp lệ."

	// Device status messages
	case status.DEVICE_INVALID_PARAMS:
		return "Tham số thiết bị không hợp lệ."
	case status.DEVICE_MISSING_UUID:
		return "Thiếu UUID thiết bị."

	// OTP status messages
	case status.OTP_MISSING_PURPOSE:
		return "Thiếu mục đích OTP."
	case status.OTP_INVALID_PURPOSE:
		return "Mục đích OTP không hợp lệ."
	case status.OTP_MISSING_IDENTIFIER:
		return "Thiếu định danh OTP."
	case status.OTP_MISSING_CODE:
		return "Thiếu mã OTP."
	case status.OTP_INVALID_CODE:
		return "Mã OTP không hợp lệ."
	case status.OTP_STILL_ACTIVE:
		return "OTP vẫn còn hiệu lực."
	case status.OTP_EXCEED_MAX_SEND:
		return "Vượt quá số lần gửi OTP tối đa."
	case status.OTP_EXCEED_MAX_VERIFY:
		return "Vượt quá số lần xác minh OTP tối đa."
	case status.OTP_EXPIRED:
		return "OTP đã hết hạn."
	case status.OTP_NOT_ALLOWED:
		return "Không cho phép sử dụng OTP."
	case status.OTP_BLOCK_DEVICE:
		return "Thiết bị đã bị chặn do vi phạm OTP."
	case status.OTP_BLOCK_DEVICE_PHONE:
		return "Số điện thoại thiết bị đã bị chặn do vi phạm OTP."
	case status.OTP_BLOCK_DEVICE_EMAIL:
		return "Email thiết bị đã bị chặn do vi phạm OTP."

		// Grade status messages
	case status.GRADE_INVALID_PARAMS:
		return "Tham số cấp học không hợp lệ."
	case status.GRADE_MISSING_ID:
		return "Thiếu ID cấp học."
	case status.GRADE_MISSING_LABEL:
		return "Thiếu tên cấp học."
	case status.GRADE_NOT_FOUND:
		return "Không tìm thấy cấp học."
	case status.GRADE_ALREADY_EXISTS:
		return "Cấp học đã tồn tại."
	case status.GRADE_CANNOT_DELETE:
		return "Cấp học không thể bị xóa."

	// Level status messages
	case status.LEVEL_INVALID_PARAMS:
		return "Tham số cấp độ không hợp lệ."
	case status.LEVEL_MISSING_LABEL:
		return "Thiếu tên cấp độ."
	case status.LEVEL_NOT_FOUND:
		return "Không tìm thấy cấp độ."
	case status.LEVEL_ALREADY_EXISTS:
		return "Cấp độ đã tồn tại."
	case status.LEVEL_CANNOT_DELETE:
		return "Cấp độ không thể bị xóa."

	// Profile status messages
	case status.PROFILE_INVALID_PARAMS:
		return "Tham số hồ sơ không hợp lệ."
	case status.PROFILE_MISSING_UID:
		return "Thiếu ID người dùng trong hồ sơ."
	case status.PROFILE_MISSING_GRADE:
		return "Thiếu cấp học trong hồ sơ."
	case status.PROFILE_MISSING_SEMESTER:
		return "Thiếu cấp độ trong hồ sơ."
	case status.PROFILE_NOT_FOUND:
		return "Không tìm thấy hồ sơ."
	case status.PROFILE_ALREADY_EXISTS:
		return "Hồ sơ đã tồn tại."
	case status.PROFILE_CANNOT_DELETE:
		return "Hồ sơ không thể bị xóa."
	case status.PROFILE_INVALID_GRADE:
		return "Cấp học trong hồ sơ không hợp lệ."
	case status.PROFILE_INVALID_LEVEL:
		return "Cấp độ trong hồ sơ không hợp lệ."

	// Semester status messages
	case status.SEMESTER_INVALID_PARAMS:
		return "Tham số học kỳ không hợp lệ."
	case status.SEMESTER_MISSING_ID:
		return "Thiếu ID học kỳ."
	case status.SEMESTER_MISSING_NAME:
		return "Thiếu tên học kỳ."
	case status.SEMESTER_NOT_FOUND:
		return "Không tìm thấy học kỳ."
	case status.SEMESTER_ALREADY_EXISTS:
		return "Học kỳ đã tồn tại."
	case status.SEMESTER_CANNOT_DELETE:
		return "Học kỳ không thể bị xóa."

	default:
		return "Unknown"
	}
}
