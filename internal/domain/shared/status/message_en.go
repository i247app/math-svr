package status

func GetENMessage(statusCode StatusCode) StatusMessage {
	switch statusCode {
	case SUCCESS:
		return "Success"
	case BAD_REQUEST:
		return "Bad Request"
	case UNAUTHORIZED:
		return "Unauthorized"
	case FORBIDDEN:
		return "Forbidden"
	case NOT_FOUND:
		return "Not Found"
	case CONFLICT:
		return "Conflict"
	case INTERNAL_SERVER_ERROR:
		return "Internal Server Error"

	// User
	case USER_MISSING_EMAIL:
		return "Please enter your email"
	case USER_MISSING_PHONE:
		return "Please enter your phone"
	case USER_MISSING_PASSWORD:
		return "Please enter your password"
	case USER_INVALID_EMAIL:
		return "User invalid email"
	case USER_INVALID_PASSWORD:
		return "User invalid password"
	case USER_EMAIL_ALREADY_EXISTS:
		return "Email already exists"
	case USER_PHONE_ALREADY_EXISTS:
		return "Phone already exists"
	case USER_NOT_FOUND:
		return "User not found"

	// Auth
	case AUTH_MISSING_PHONE:
		return "Please enter your phone"
	case AUTH_MISSING_DEVICE_UUID:
		return "Device identifier is required"
	case AUTH_MISSING_IP_ADDRESS:
		return "IP address is required"
	case AUTH_MISSING_DEVICE_PUSH_TOKEN:
		return "Device push token is required"
	case AUTH_INVALID_TOKEN:
		return "Invalid or expired token"
	case AUTH_LOGIN_FAILED:
		return "Login failed"
	case AUTH_LOGOUT_FAILED:
		return "Logout failed"

	// Profile
	case PROFILE_NOT_FOUND:
		return "Profile not found"
	case PROFILE_MISSING_NAME:
		return "Please enter the child's name"
	case PROFILE_MISSING_USER_ID:
		return "User id is required"
	case PROFILE_MISSING_GRADE_ID:
		return "Grade id is required"
	case PROFILE_MISSING_SEMESTER_ID:
		return "Semester id is required"
	case PROFILE_INVALID_DOB:
		return "Date of birth is invalid"
	case PROFILE_AVATAR_INVALID_FILE:
		return "Avatar file is invalid"
	case PROFILE_AVATAR_UPLOAD_FAILED:
		return "Avatar upload failed"
	case PROFILE_MISSING_PROGRAM_ID:
		return "Program id is required"
	case PROFILE_INVALID_LANGUAGE:
		return "Language is invalid"

	// Program
	case PROGRAM_NOT_FOUND:
		return "Program not found"
	case PROGRAM_INVALID_LANGUAGE:
		return "Program language is invalid"

	// Grade
	case GRADE_NOT_FOUND:
		return "Grade not found"
	case GRADE_INVALID_LANGUAGE:
		return "Grade language is invalid"

	// Semester
	case SEMESTER_NOT_FOUND:
		return "Semester not found"
	case SEMESTER_INVALID_LANGUAGE:
		return "Semester language is invalid"

	// Device
	case DEVICE_NOT_FOUND:
		return "Device not found"
	case DEVICE_MISSING_UUID:
		return "Device identifier is required"
	case DEVICE_MISSING_NAME:
		return "Device name is required"
	case DEVICE_MISSING_USER_ID:
		return "User id is required"
	case DEVICE_NOT_OWNED:
		return "Device does not belong to this user"
	case DEVICE_ALREADY_VERIFIED:
		return "Device is already verified"
	case DEVICE_REGISTRATION_FAIL:
		return "Device registration failed"
	case DEVICE_VERIFICATION_FAIL:
		return "Device verification failed"
	case DEVICE_REVOKE_FAIL:
		return "Device revoke failed"
	case DEVICE_2FA_REQUIRED:
		return "Two-factor authentication is required for this device"

	// Notification
	case NOTIFICATION_MISSING_UID:
		return "User id is required"
	case NOTIFICATION_MISSING_TITLE:
		return "Notification title is required"
	case NOTIFICATION_MISSING_SHORT_TEXT:
		return "Notification short text is required"
	case NOTIFICATION_INVALID_ICON:
		return "Notification icon is invalid"
	case NOTIFICATION_INVALID_PRIORITY:
		return "Notification priority is invalid"
	case NOTIFICATION_NOT_FOUND:
		return "Notification not found"

	// Messaging
	case MESSAGING_PUBLISH_FAILED:
		return "Failed to publish event"
	case MESSAGING_INVALID_ENVELOPE:
		return "Event envelope is invalid"
	case MESSAGING_CONFIG_INVALID:
		return "Messaging configuration is invalid"
	case MESSAGING_CONNECT_FAILED:
		return "Failed to connect to messaging broker"

	// Cache
	case CACHE_CONNECT_FAILED:
		return "Failed to connect to cache"
	case CACHE_CONFIG_INVALID:
		return "Cache configuration is invalid"
	case CACHE_OP_FAILED:
		return "Cache operation failed"
	case CACHE_SERIALIZE_FAILED:
		return "Failed to serialize cache value"

	// SMS
	case SMS_CONNECT_FAILED:
		return "SMS provider connection failed"
	case SMS_CONFIG_INVALID:
		return "SMS configuration is invalid"
	case SMS_OP_FAILED:
		return "SMS operation failed"
	case SMS_SERIALIZE_FAILED:
		return "Failed to serialize SMS payload"

	// Bot (AI / chat)
	case BOT_CONNECT_FAILED:
		return "AI provider connection failed"
	case BOT_CONFIG_INVALID:
		return "AI configuration is invalid"
	case BOT_OP_FAILED:
		return "AI operation failed"
	case BOT_SERIALIZE_FAILED:
		return "Failed to decode AI response"
	case BOT_INVALID_PROMPT:
		return "AI prompt is invalid"
	case BOT_CONTEXT_TOO_LARGE:
		return "AI context window exceeded"
	case BOT_RATE_LIMITED:
		return "AI provider rate limit exceeded"
	case BOT_UNSUPPORTED_OP:
		return "AI operation is not supported by the configured provider"
	default:
		return ""
	}
}
