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
	case NO_DATA:
		return "No Resource found"
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
	case USER_USERNAME_ALREADY_EXISTS:
		return "Username already exists"
	case USER_NOT_FOUND:
		return "User not found"
	case USER_MISSING_NAME:
		return "Please enter your name"
	case USER_AVATAR_INVALID_FILE:
		return "Avatar file is invalid"
	case USER_AVATAR_UPLOAD_FAILED:
		return "Avatar upload failed"
	case USER_AVATAR_CONFLICT:
		return "Provide either an avatar file or an avatar reference, not both"
	case USER_AVATAR_INVALID_REFERENCE:
		return "Avatar reference is invalid"

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
	case PROFILE_AVATAR_CONFLICT:
		return "Provide either an avatar file or an avatar reference, not both"
	case PROFILE_AVATAR_INVALID_REFERENCE:
		return "Avatar reference is invalid"
	case PROFILE_MISSING_PROGRAM_ID:
		return "Program id is required"
	case PROFILE_INVALID_LANGUAGE:
		return "Language is invalid"

	// Program
	case PROGRAM_NOT_FOUND:
		return "Program not found"
	case PROGRAM_INVALID_LANGUAGE:
		return "Program language is invalid"
	case PROGRAM_MISSING_ID:
		return "Program id is required"
	case PROGRAM_MISSING_LABEL:
		return "Program label is required"
	case PROGRAM_MISSING_DESCRIPTION:
		return "Program description is required"
	case PROGRAM_INVALID_DISPLAY_ORDER:
		return "Program display order is invalid"
	case PROGRAM_INVALID_TRANSLATION:
		return "Program translation payload is invalid"
	case PROGRAM_TRANSLATION_ALREADY_EXISTS:
		return "A translation in this language already exists for the program"
	case PROGRAM_TRANSLATION_NOT_FOUND:
		return "Program translation not found"

	// Grade
	case GRADE_NOT_FOUND:
		return "Grade not found"
	case GRADE_INVALID_LANGUAGE:
		return "Grade language is invalid"
	case GRADE_MISSING_ID:
		return "Grade id is required"
	case GRADE_MISSING_LABEL:
		return "Grade label is required"
	case GRADE_MISSING_DESCRIPTION:
		return "Grade description is required"
	case GRADE_INVALID_DISPLAY_ORDER:
		return "Grade display order is invalid"
	case GRADE_INVALID_TRANSLATION:
		return "Grade translation payload is invalid"
	case GRADE_TRANSLATION_ALREADY_EXISTS:
		return "A translation in this language already exists for the grade"
	case GRADE_TRANSLATION_NOT_FOUND:
		return "Grade translation not found"

	// Semester
	case SEMESTER_NOT_FOUND:
		return "Semester not found"
	case SEMESTER_INVALID_LANGUAGE:
		return "Semester language is invalid"
	case SEMESTER_MISSING_ID:
		return "Semester id is required"
	case SEMESTER_MISSING_NAME:
		return "Semester name is required"
	case SEMESTER_MISSING_DESCRIPTION:
		return "Semester description is required"
	case SEMESTER_INVALID_DISPLAY_ORDER:
		return "Semester display order is invalid"
	case SEMESTER_INVALID_TRANSLATION:
		return "Semester translation payload is invalid"
	case SEMESTER_TRANSLATION_ALREADY_EXISTS:
		return "A translation in this language already exists for the semester"
	case SEMESTER_TRANSLATION_NOT_FOUND:
		return "Semester translation not found"

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

	// School
	case SCHOOL_NOT_FOUND:
		return "School not found"
	case SCHOOL_MISSING_ID:
		return "School id is required"
	case SCHOOL_MISSING_NAME:
		return "School name is required"
	case SCHOOL_NAME_TOO_LONG:
		return "School name is too long"
	case SCHOOL_INVALID_STATUS:
		return "School status is invalid"
	case SCHOOL_DESCRIPTION_TOO_LONG:
		return "School description is too long"
	case SCHOOL_PROFILE_LINK_FAILED:
		return "Failed to link profile with school"

	// Chapter
	case CHAPTER_NOT_FOUND:
		return "Chapter not found"
	case CHAPTER_MISSING_ID:
		return "Chapter id is required"
	case CHAPTER_MISSING_PROGRAM_ID:
		return "Program id is required"
	case CHAPTER_MISSING_GRADE_ID:
		return "Grade id is required"
	case CHAPTER_MISSING_SEMESTER_ID:
		return "Semester id is required"
	case CHAPTER_MISSING_LABEL:
		return "Chapter label is required"
	case CHAPTER_MISSING_DESCRIPTION:
		return "Chapter description is required"
	case CHAPTER_INVALID_DISPLAY_ORDER:
		return "Chapter display order is invalid"
	case CHAPTER_INVALID_LANGUAGE:
		return "Chapter language is invalid"
	case CHAPTER_INVALID_TRANSLATION:
		return "Chapter translation payload is invalid"
	case CHAPTER_TRANSLATION_ALREADY_EXISTS:
		return "A translation in this language already exists for the chapter"
	case CHAPTER_TRANSLATION_NOT_FOUND:
		return "Chapter translation not found"

	// OTP
	case OTP_NOT_FOUND:
		return "OTP not found"
	case OTP_MISSING_TYPE:
		return "OTP type is required"
	case OTP_MISSING_IDENTIFIER:
		return "Phone or email is required"
	case OTP_MISSING_CODE:
		return "OTP code is required"
	case OTP_INVALID_TYPE:
		return "OTP type is invalid"
	case OTP_INVALID_CODE:
		return "OTP code is invalid"
	case OTP_EXPIRED:
		return "OTP has expired"
	case OTP_ALREADY_VERIFIED:
		return "OTP has already been verified"
	case OTP_REVOKED:
		return "OTP has been revoked"
	case OTP_TOO_MANY_ATTEMPTS:
		return "Too many verification attempts; request a new OTP"
	case OTP_TOO_FREQUENT:
		return "Please wait before requesting another OTP"
	case OTP_RATE_LIMITED:
		return "OTP request limit reached; try again later"
	case OTP_DELIVERY_FAILED:
		return "OTP delivery failed"
	case OTP_NO_DELIVERY_CHANNEL:
		return "No delivery channel is configured for this identifier"
	case OTP_GENERATION_FAILED:
		return "Failed to generate OTP"
	case OTP_DISABLED:
		return "OTP is disabled"

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

	// Quiz
	case QUIZ_NOT_FOUND:
		return "Quiz not found"
	case QUIZ_MISSING_PROFILE_ID:
		return "Profile id is required"
	case QUIZ_MISSING_TYPE:
		return "Quiz type is required"
	case QUIZ_INVALID_TYPE:
		return "Quiz type is invalid"
	case QUIZ_INVALID_LANGUAGE:
		return "Quiz language is invalid"
	case QUIZ_MISSING_ANSWERS:
		return "Quiz answers are required"
	case QUIZ_INVALID_ANSWERS:
		return "Quiz answers are invalid"
	case QUIZ_ALREADY_SUBMITTED:
		return "This quiz has already been submitted"
	case QUIZ_NOT_OWNED:
		return "Quiz does not belong to this profile"
	case QUIZ_PREVIOUS_NOT_FOUND:
		return "Previous quiz not found"
	case QUIZ_PREVIOUS_NOT_GRADED:
		return "Previous quiz has not been graded yet"
	case QUIZ_GENERATION_FAILED:
		return "Failed to generate quiz questions"
	case QUIZ_GRADING_FAILED:
		return "Failed to grade quiz answers"
	case QUIZ_PROFILE_NOT_CONFIGURED:
		return "Profile must have a program, grade, and semester set before generating a quiz"
	case QUIZ_INVALID_TYPE_OF_QUIZ:
		return "Quiz learning type is invalid"

	// Job
	case JOB_NOT_FOUND:
		return "Job not found"
	case JOB_ALREADY_PAUSED:
		return "Job is already paused"
	case JOB_NOT_PAUSED:
		return "Job is not paused"
	case JOB_IN_FLIGHT:
		return "Job is already running an execution"
	case JOB_RUNTIME_UNAVAILABLE:
		return "Job runtime is not started or is shutting down"
	case JOB_MISSING_NAME:
		return "Job name is required"
	case JOB_TRIGGER_FAILED:
		return "Failed to trigger job"
	case TASK_HANDLER_NOT_FOUND:
		return "Task handler is not registered"
	case TASK_QUEUE_FULL:
		return "Task queue is full; please retry shortly"
	case TASK_MISSING_NAME:
		return "Task name is required"
	case TASK_ENQUEUE_FAILED:
		return "Failed to enqueue task"

	// Sequence
	case SEQ_NOT_FOUND:
		return "Sequence not found"
	case SEQ_MISSING_NAME:
		return "Sequence name is required"
	case SEQ_GENERATION_FAILED:
		return "Failed to generate next id"

	// Classroom
	case CLASSROOM_NOT_FOUND:
		return "Classroom not found"
	case CLASSROOM_MISSING_ID:
		return "Classroom id is required"
	case CLASSROOM_MISSING_NAME:
		return "Classroom name is required"
	case CLASSROOM_NAME_TOO_LONG:
		return "Classroom name is too long"
	case CLASSROOM_DESCRIPTION_TOO_LONG:
		return "Classroom description is too long"
	case CLASSROOM_MISSING_OWNER_PROFILE_ID:
		return "Owner profile id is required"
	case CLASSROOM_INVALID_OWNER_ROLE:
		return "Only a teacher profile can own a classroom"
	case CLASSROOM_INVALID_PROGRAM:
		return "Program is invalid"
	case CLASSROOM_INVALID_GRADE:
		return "Grade is invalid"
	case CLASSROOM_ALREADY_ARCHIVED:
		return "Classroom is already archived"
	case CLASSROOM_NOT_ARCHIVED:
		return "Classroom is not archived"
	case CLASSROOM_ALREADY_DELETED:
		return "Classroom is already deleted"
	case CLASSROOM_MAX_MEMBERS_REACHED:
		return "Classroom has reached its member limit"
	case CLASSROOM_CODE_INVALID:
		return "Invite code is invalid"
	case CLASSROOM_CODE_EXPIRED:
		return "Invite code has expired"
	case CLASSROOM_CODE_DISABLED:
		return "Invite code is disabled"
	case CLASSROOM_CODE_GENERATION_FAILED:
		return "Failed to generate invite code"
	case CLASSROOM_PERMISSION_DENIED:
		return "You do not have permission to perform this action on the classroom"
	case CLASSROOM_OWNER_CANNOT_LEAVE:
		return "Owner must transfer ownership before leaving the classroom"
	case CLASSROOM_OWNER_TRANSFER_TO_NON_MEMBER:
		return "Ownership can only be transferred to an active member"
	case CLASSROOM_INVALID_MAX_MEMBERS:
		return "Max members must be a positive number"
	case CLASSROOM_INVALID_SCHOOL:
		return "School not found"
	case CLASSROOM_CODE_TAKEN:
		return "Invite code is already in use"
	case CLASSROOM_PROGRAM_DUPLICATE:
		return "Duplicate program id in the list"

	// Classroom member
	case CLASSROOM_MEMBER_NOT_FOUND:
		return "Classroom member not found"
	case CLASSROOM_MEMBER_MISSING_ID:
		return "Member id is required"
	case CLASSROOM_MEMBER_MISSING_CLASSROOM_ID:
		return "Classroom id is required"
	case CLASSROOM_MEMBER_MISSING_PROFILE_ID:
		return "Profile id is required"
	case CLASSROOM_MEMBER_ALREADY_MEMBER:
		return "Profile is already a member of this classroom"
	case CLASSROOM_MEMBER_NOT_MEMBER:
		return "Profile is not a member of this classroom"
	case CLASSROOM_MEMBER_INVALID_ROLE:
		return "Member role is invalid"
	case CLASSROOM_MEMBER_CANNOT_REMOVE_OWNER:
		return "Owner cannot be removed"
	case CLASSROOM_MEMBER_CANNOT_DEMOTE_OWNER:
		return "Owner role cannot be changed; transfer ownership first"

	// Classroom invitation
	case CLASSROOM_INVITATION_NOT_FOUND:
		return "Invitation not found"
	case CLASSROOM_INVITATION_MISSING_ID:
		return "Invitation id is required"
	case CLASSROOM_INVITATION_MISSING_TARGET:
		return "Invitation must target a profile or an identifier"
	case CLASSROOM_INVITATION_INVALID_IDENTIFIER:
		return "Invitation identifier is invalid"
	case CLASSROOM_INVITATION_INVALID_IDENTIFIER_TYPE:
		return "Invitation identifier type is invalid"
	case CLASSROOM_INVITATION_INVALID_ROLE:
		return "Proposed role is invalid"
	case CLASSROOM_INVITATION_ALREADY_RESPONDED:
		return "Invitation has already been responded to"
	case CLASSROOM_INVITATION_EXPIRED:
		return "Invitation has expired"
	case CLASSROOM_INVITATION_NOT_PENDING:
		return "Invitation is not in a pending state"
	case CLASSROOM_INVITATION_NOT_INVITEE:
		return "You are not the invitee of this invitation"
	case CLASSROOM_INVITATION_ALREADY_INVITED:
		return "A pending invitation already exists for this invitee"
	case CLASSROOM_INVITATION_TOKEN_INVALID:
		return "Invitation token is invalid"
	case CLASSROOM_INVITATION_TOKEN_GENERATION_FAILED:
		return "Failed to generate invitation token"
	case CLASSROOM_INVITATION_PERMISSION_DENIED:
		return "You do not have permission to manage this invitation"

	// Classroom join request
	case CLASSROOM_JOIN_REQUEST_NOT_FOUND:
		return "Join request not found"
	case CLASSROOM_JOIN_REQUEST_NOT_PENDING:
		return "Join request is not in a pending state"
	case CLASSROOM_JOIN_REQUEST_ALREADY_PENDING:
		return "A pending join request already exists for this classroom"
	case CLASSROOM_JOIN_REQUEST_PERMISSION_DENIED:
		return "You do not have permission to manage this join request"

	// Classroom exercise
	case CLASSROOM_EXERCISE_NOT_FOUND:
		return "Classroom exercise not found"
	case CLASSROOM_EXERCISE_MISSING_ID:
		return "Classroom exercise id is required"
	case CLASSROOM_EXERCISE_MISSING_CLASSROOM_ID:
		return "Classroom id is required"
	case CLASSROOM_EXERCISE_MISSING_TITLE:
		return "Exercise title is required"
	case CLASSROOM_EXERCISE_TITLE_TOO_LONG:
		return "Exercise title is too long"
	case CLASSROOM_EXERCISE_MISSING_CHAPTER_NAME:
		return "Chapter name is required"
	case CLASSROOM_EXERCISE_CHAPTER_NAME_TOO_LONG:
		return "Chapter name is too long"
	case CLASSROOM_EXERCISE_MISSING_LESSON_NAME:
		return "Lesson name is required"
	case CLASSROOM_EXERCISE_LESSON_NAME_TOO_LONG:
		return "Lesson name is too long"
	case CLASSROOM_EXERCISE_INVALID_NUM_QUESTIONS:
		return "Number of questions is invalid"
	case CLASSROOM_EXERCISE_INVALID_DATE_RANGE:
		return "Exercise end date must be after start date"
	case CLASSROOM_EXERCISE_NOTE_TOO_LONG:
		return "Exercise note is too long"
	case CLASSROOM_EXERCISE_INVALID_PROGRAM:
		return "Program is invalid"
	case CLASSROOM_EXERCISE_PROGRAM_NOT_IN_CLASSROOM:
		return "Program is not associated with this classroom"
	case CLASSROOM_EXERCISE_ALREADY_DELETED:
		return "Classroom exercise is already deleted"
	case CLASSROOM_EXERCISE_PERMISSION_DENIED:
		return "You do not have permission to manage this exercise"
	case CLASSROOM_EXERCISE_GENERATION_FAILED:
		return "Failed to generate classroom exercise"
	case CLASSROOM_EXERCISE_INVALID_VISIBILITY:
		return "Visibility must be PUBLIC or PRIVATE"
	case CLASSROOM_EXERCISE_PRIVATE_DENIED:
		return "Only the creator can access this private exercise"
	default:
		return ""
	}
}
