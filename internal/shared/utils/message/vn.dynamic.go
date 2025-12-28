package message

import "math-ai.com/math-ai/internal/shared/constant/status"

// GetMessageTemplateVN returns Vietnamese message templates with placeholders for dynamic arguments
// If a template exists for the status code, it returns the template; otherwise returns empty string
// Use this when you have dynamic arguments to interpolate
func GetMessageTemplateVN(statusCode status.Code) string {
	switch statusCode {
	// User templates with dynamic arguments
	case status.USER_EMAIL_ALREADY_EXISTS:
		return "Email {email} đã tồn tại."
	case status.USER_PHONE_ALREADY_EXISTS:
		return "Số điện thoại {phone} đã tồn tại."
	case status.USER_NOT_FOUND:
		return "Không tìm thấy người dùng có ID {user_id}."

	// Grade templates
	case status.GRADE_ALREADY_EXISTS:
		return "Cấp học '{label}' đã tồn tại."
	case status.GRADE_NOT_FOUND:
		return "Không tìm thấy cấp học '{label}'."

	// Semester templates
	case status.SEMESTER_ALREADY_EXISTS:
		return "Học kỳ '{name}' đã tồn tại."
	case status.SEMESTER_NOT_FOUND:
		return "Không tìm thấy học kỳ '{name}'."

	// Level templates
	case status.LEVEL_ALREADY_EXISTS:
		return "Cấp độ '{label}' đã tồn tại."
	case status.LEVEL_NOT_FOUND:
		return "Không tìm thấy cấp độ '{label}'."

	// Contact templates
	case status.CONTACT_NAME_TOO_LONG:
		return "Tên liên hệ quá dài (tối đa {max_length} ký tự)."
	case status.CONTACT_MESSAGE_TOO_LONG:
		return "Nội dung liên hệ quá dài (tối đa {max_length} ký tự)."
	case status.CONTACT_EMAIL_TOO_LONG:
		return "Email liên hệ quá dài (tối đa {max_length} ký tự)."

	default:
		return "" // No template available, use static message
	}
}
