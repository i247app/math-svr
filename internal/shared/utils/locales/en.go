package locales

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

var (
	EN LanguageType = "en"
)

func GetMessageENFromStatus(statusCode status.Code, args ...any) string {
	switch statusCode {
	case status.USER_INVALID_PARAMS:
		return "Invalid parameters"
	case status.USER_INVALID_ID:
		return "Invalid user ID"
	case status.USER_NOT_FOUND:
		return "User not found"
	case status.USER_MISSING_FIRST_NAME:
		return "First name is required"
	case status.USER_MISSING_LAST_NAME:
		return "Last name is required"
	case status.USER_MISSING_EMAIL:
		return "Email is required"
	case status.USER_MISSING_PASSWORD:
		return "Password is required"
	case status.USER_INVALID_EMAIL:
		return "Invalid email format"
	case status.USER_EMAIL_ALREADY_EXISTS:
		return "Email already exists"
	case status.USER_MISSING_PHONE:
		return "Phone is required"
	case status.USER_INVALID_PHONE:
		return "Invalid phone format"
	case status.USER_PHONE_ALREADY_EXISTS:
		return "Phone already exists"
	case status.USER_INVALID_ROLE:
		return fmt.Sprintf("Invalid role. Valid roles are: %v", args)
	case status.USER_INVALID_STATUS:
		return fmt.Sprintf("Invalid status. Valid statuses are: %v", args)
	case status.DEVICE_INVALID_PARAMS:
		return "Invalid device parameters"
	case status.DEVICE_MISSING_UUID:
		return "Device UUID is required"
	case status.DEVICE_BLOCKED:
		return "Device is blocked"
	case status.DEVICE_MISSING_NAME:
		return "Device name is required"
	case status.LOGIN_MISSING_PARAMETERS:
		return "Missing required parameters"
	case status.LOGIN_WRONG_CREDENTIALS:
		return "Wrong login credentials"
	case status.BLOCK_MISSING_TYPE:
		return "Block type is required"
	case status.BLOCK_INVALID_TYPE:
		return fmt.Sprintf("Invalid block type. Valid statuses are: %v", args)
	case status.BLOCK_MISSING_VALUE:
		return "Block value is required"
	case status.OTP_MISSING_PURPOSE:
		return "OTP purpose is required"
	case status.OTP_INVALID_PURPOSE:
		return fmt.Sprintf("Invalid OTP purpose. Valid purposes are: %v", args)
	case status.OTP_MISSING_IDENTIFIER:
		return "OTP identifier is required"
	case status.OTP_MISSING_CODE:
		return "OTP code is required"
	case status.OTP_INVALID_CODE:
		return "Invalid OTP code"
	case status.OTP_STILL_ACTIVE:
		return fmt.Sprintf("OTP is still active, please try again after %d seconds", args...)
	case status.OTP_EXCEED_MAX_SEND:
		return "Exceeded maximum OTP send attempts"
	case status.OTP_EXCEED_MAX_VERIFY:
		return fmt.Sprintf("Exceeded maximum OTP verify attempts, wait %d seconds to requeset otp again", args...)
	case status.OTP_EXPIRED:
		return "OTP has expired"
	case status.OTP_NOT_ALLOWED:
		return "OTP action not allowed"
	case status.OTP_BLOCK_DEVICE:
		return fmt.Sprintf("For security reasons, this device has been blocked within %d minutes", args...)
	case status.OTP_BLOCK_DEVICE_PHONE:
		return fmt.Sprintf("For security reasons, this device and phone number has been blocked within %d minutes", args...)
	case status.OTP_BLOCK_DEVICE_EMAIL:
		return fmt.Sprintf("For security reasons, this device and email has been blocked within %d minutes", args...)
	case status.SUCCESS:
		return "Success"
	default:
		return "Unknown"
	}
}
