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
	default:
		return ""
	}
}
