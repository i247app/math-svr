package message

import "math-ai.com/math-ai/internal/shared/constant/status"

// GetMessageTemplateEN returns message templates with placeholders for dynamic arguments
// If a template exists for the status code, it returns the template; otherwise returns empty string
// Use this when you have dynamic arguments to interpolate
func GetMessageTemplateEN(statusCode status.Code) string {
	switch statusCode {
	// User templates with dynamic arguments
	case status.USER_EMAIL_ALREADY_EXISTS:
		return "Email {email} already exists."
	case status.USER_PHONE_ALREADY_EXISTS:
		return "Phone number {phone} already exists."
	case status.USER_NOT_FOUND:
		return "User with ID {user_id} not found."

	// Grade templates
	case status.GRADE_ALREADY_EXISTS:
		return "Grade '{label}' already exists."
	case status.GRADE_NOT_FOUND:
		return "Grade '{label}' not found."

	// Semester templates
	case status.SEMESTER_ALREADY_EXISTS:
		return "Semester '{name}' already exists."
	case status.SEMESTER_NOT_FOUND:
		return "Semester '{name}' not found."

	// Level templates
	case status.LEVEL_ALREADY_EXISTS:
		return "Level '{label}' already exists."
	case status.LEVEL_NOT_FOUND:
		return "Level '{label}' not found."

	// Contact templates
	case status.CONTACT_NAME_TOO_LONG:
		return "Contact name is too long (max {max_length} characters)."
	case status.CONTACT_MESSAGE_TOO_LONG:
		return "Contact message is too long (max {max_length} characters)."
	case status.CONTACT_EMAIL_TOO_LONG:
		return "Contact email is too long (max {max_length} characters)."

	default:
		return "" // No template available, use static message
	}
}
