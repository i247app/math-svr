package locales

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

var (
	EN LanguageType = "en"
)

func GetMessageENFromStatus(statusCode status.Code) string {
	args := GetArgsByStatatus(statusCode)

	switch statusCode {
	case status.OK:
		return "Success"
	case status.CREATED:
		return "Resource created successfully."
	case status.FAIL:
		return "Request failed."
	case status.UNAUTHORIZED:
		return "Unauthorized access."
	case status.NOT_FOUND:
		return "Resource not found."
	case status.INTERNAL:
		return "Internal server error."

	// User status messages
	case status.USER_MISSING_ID:
		return "User ID is missing."
	case status.USER_INVALID_PARAMS:
		return "Invalid user parameters."
	case status.USER_INVALID_ID:
		return "Invalid user ID."
	case status.USER_NOT_FOUND:
		return "User not found."
	case status.USER_MISSING_NAME:
		return "User name is missing."
	case status.USER_MISSING_EMAIL:
		return "User email is missing."
	case status.USER_INVALID_EMAIL:
		return "Invalid user email format."
	case status.USER_EMAIL_ALREADY_EXISTS:
		return "Email already exists."
	case status.USER_MISSING_PHONE:
		return "User phone number is missing."
	case status.USER_PHONE_ALREADY_EXISTS:
		return "Phone number already exists."
	case status.USER_INVALID_PHONE:
		return "Invalid user phone number."
	case status.USER_INVALID_ROLE:
		return fmt.Sprintf("Invalid role. Valid roles are: %v", args)
	case status.USER_INVALID_STATUS:
		return "Invalid user status."

	// Device status messages
	case status.DEVICE_INVALID_PARAMS:
		return "Invalid device parameters."
	case status.DEVICE_MISSING_UUID:
		return "Device UUID is missing."
	case status.DEVICE_MISSING_NAME:
		return "Device name is missing."
	case status.DEVICE_BLOCKED:
		return "Device is blocked."

	// Login status messages
	case status.LOGIN_MISSING_PARAMETERS:
		return "Missing login parameters."
	case status.LOGIN_WRONG_CREDENTIALS:
		return "Wrong login credentials."

	// Block status messages
	case status.BLOCK_MISSING_TYPE:
		return "Block type is missing."
	case status.BLOCK_MISSING_VALUE:
		return "Block value is missing."
	case status.BLOCK_INVALID_TYPE:
		return "Invalid block type."

	// OTP status messages
	case status.OTP_MISSING_PURPOSE:
		return "OTP purpose is missing."
	case status.OTP_INVALID_PURPOSE:
		return "Invalid OTP purpose."
	case status.OTP_MISSING_IDENTIFIER:
		return "OTP identifier is missing."
	case status.OTP_MISSING_CODE:
		return "OTP code is missing."
	case status.OTP_INVALID_CODE:
		return "Invalid OTP code."
	case status.OTP_STILL_ACTIVE:
		return "An active OTP already exists."
	case status.OTP_EXCEED_MAX_SEND:
		return "Exceeded maximum OTP send attempts."
	case status.OTP_EXCEED_MAX_VERIFY:
		return "Exceeded maximum OTP verify attempts."
	case status.OTP_EXPIRED:
		return "OTP has expired."
	case status.OTP_NOT_ALLOWED:
		return "OTP request not allowed."
	case status.OTP_BLOCK_DEVICE:
		return "Device is blocked due to OTP violations."
	case status.OTP_BLOCK_DEVICE_PHONE:
		return "Device phone is blocked due to OTP violations."
	case status.OTP_BLOCK_DEVICE_EMAIL:
		return "Device email is blocked due to OTP violations."

	// Grade status messages
	case status.GRADE_INVALID_PARAMS:
		return "Invalid grade parameters."
	case status.GRADE_MISSING_ID:
		return "Grade ID is missing."
	case status.GRADE_MISSING_LABEL:
		return "Grade name is missing."
	case status.GRADE_NOT_FOUND:
		return "Grade not found."
	case status.GRADE_ALREADY_EXISTS:
		return "Grade already exists."
	case status.GRADE_CANNOT_DELETE:
		return "Grade cannot be deleted."

	// Level status messages
	case status.LEVEL_INVALID_PARAMS:
		return "Invalid level parameters."
	case status.LEVEL_MISSING_LABEL:
		return "Level name is missing."
	case status.LEVEL_NOT_FOUND:
		return "Level not found."
	case status.LEVEL_ALREADY_EXISTS:
		return "Level already exists."
	case status.LEVEL_CANNOT_DELETE:
		return "Level cannot be deleted."

	// Profile status messages
	case status.PROFILE_INVALID_PARAMS:
		return "Invalid profile parameters."
	case status.PROFILE_MISSING_UID:
		return "Profile user ID is missing."
	case status.PROFILE_MISSING_GRADE:
		return "Profile grade is missing."
	case status.PROFILE_MISSING_SEMESTER:
		return "Profile level is missing."
	case status.PROFILE_NOT_FOUND:
		return "Profile not found."
	case status.PROFILE_ALREADY_EXISTS:
		return "Profile already exists."
	case status.PROFILE_CANNOT_DELETE:
		return "Profile cannot be deleted."
	case status.PROFILE_INVALID_GRADE:
		return "Invalid profile grade."
	case status.PROFILE_INVALID_LEVEL:
		return "Invalid profile level."

	// Semester status messages
	case status.SEMESTER_INVALID_PARAMS:
		return "Invalid semester parameters."
	case status.SEMESTER_MISSING_ID:
		return "Semester ID is missing."
	case status.SEMESTER_MISSING_NAME:
		return "Semester name is missing."
	case status.SEMESTER_NOT_FOUND:
		return "Semester not found."
	case status.SEMESTER_ALREADY_EXISTS:
		return "Semester already exists."
	case status.SEMESTER_CANNOT_DELETE:
		return "Semester cannot be deleted."

	default:
		return "Unknown"
	}
}
