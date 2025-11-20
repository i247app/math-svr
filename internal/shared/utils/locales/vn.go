package locales

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

var (
	VN LanguageType = "vn"
)

func GetMessageVNFromStatus(statusCode status.Code, args ...any) string {
	switch statusCode {
	case status.USER_INVALID_PARAMS:
		return "Tham số không hợp lệ"
	case status.USER_INVALID_ID:
		return "ID người dùng không hợp lệ"
	case status.USER_NOT_FOUND:
		return "Không tìm thấy người dùng"
	case status.USER_MISSING_FIRST_NAME:
		return "Thiếu tên"
	case status.USER_MISSING_LAST_NAME:
		return "Thiếu họ"
	case status.USER_MISSING_EMAIL:
		return "Thiếu email"
	case status.USER_INVALID_EMAIL:
		return "Định dạng email không hợp lệ"
	case status.USER_EMAIL_ALREADY_EXISTS:
		return "Email đã tồn tại"
	case status.USER_MISSING_PHONE:
		return "Thiếu số điện thoại"
	case status.USER_INVALID_PHONE:
		return "Định dạng số điện thoại không hợp lệ"
	case status.USER_PHONE_ALREADY_EXISTS:
		return "Số điện thoại đã tồn tại"
	case status.USER_INVALID_ROLE:
		return fmt.Sprintf("Vai trò không hợp lệ. Các vai trò hợp lệ là: %v", args)
	case status.USER_INVALID_STATUS:
		return fmt.Sprintf("Trạng thái không hợp lệ. Các trạng thái hợp lệ là: %v", args)
	case status.DEVICE_INVALID_PARAMS:
		return "Tham số thiết bị không hợp lệ"
	case status.DEVICE_MISSING_UUID:
		return "Thiếu UUID thiết bị"
	case status.DEVICE_BLOCKED:
		return "Thiết bị đã bị chặn"
	case status.DEVICE_MISSING_NAME:
		return "Thiếu tên thiết bị"
	case status.LOGIN_MISSING_PARAMETERS:
		return "Thiếu tham số bắt buộc"
	case status.LOGIN_WRONG_CREDENTIALS:
		return "Thông tin đăng nhập sai"
	case status.BLOCK_MISSING_TYPE:
		return "Thiếu loại chặn"
	case status.BLOCK_INVALID_TYPE:
		return fmt.Sprintf("Loại chặn không hợp lệ. Các trạng thái hợp lệ là: %v", args)
	case status.BLOCK_MISSING_VALUE:
		return "Thiếu giá trị chặn"
	case status.OTP_MISSING_PURPOSE:
		return "Thiếu mục đích OTP"
	case status.OTP_INVALID_PURPOSE:
		return "Mục đích OTP không hợp lệ"
	case status.OTP_MISSING_IDENTIFIER:
		return "Thiếu định danh"
	case status.OTP_MISSING_CODE:
		return "Mã OTP bị thiếu"
	case status.OTP_INVALID_CODE:
		return "Mã OTP không hợp lệ"
	case status.OTP_STILL_ACTIVE:
		return fmt.Sprintf("OTP vẫn còn hiệu lực, vui lòng thử lại sau %d giây", args...)
	case status.OTP_EXCEED_MAX_SEND:
		return "Vượt quá số lần gửi OTP tối đa"
	case status.OTP_EXCEED_MAX_VERIFY:
		return fmt.Sprintf("Vượt quá số lần xác minh OTP tối đa, vui lòng chờ %d giây để yêu cầu OTP lại", args...)
	case status.OTP_EXPIRED:
		return "OTP đã hết hạn"
	case status.OTP_NOT_ALLOWED:
		return "Hành động OTP không được phép"
	case status.OTP_BLOCK_DEVICE:
		return "Vì lý do bảo mật, thiết bị này đã bị chặn trong vòng %d phút"
	case status.OTP_BLOCK_DEVICE_PHONE:
		return "Vì lý do bảo mật, thiết bị và số điện thoại đã bị chặn trong vòng %d phút"
	case status.OTP_BLOCK_DEVICE_EMAIL:
		return "Vì lý do bảo mật, thiết bị và email đã bị chặn trong vòng %d phút"
	case status.SUCCESS:
		return "Thành công"
	default:
		return "Unknown"
	}
}
