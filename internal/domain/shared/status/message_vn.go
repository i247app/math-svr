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
	case NO_DATA:
		return "Không tìm thấy dữ liệu"
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
	case USER_USERNAME_ALREADY_EXISTS:
		return "Username đã tồn tại"
	case USER_NOT_FOUND:
		return "Không tìm thấy người dùng"
	case USER_MISSING_NAME:
		return "Vui lòng nhập tên của bạn"
	case USER_AVATAR_INVALID_FILE:
		return "Tệp ảnh đại diện không hợp lệ"
	case USER_AVATAR_UPLOAD_FAILED:
		return "Tải ảnh đại diện thất bại"

	// Auth
	case AUTH_MISSING_PHONE:
		return "Vui lòng nhập số điện thoại"
	case AUTH_MISSING_DEVICE_UUID:
		return "Vui lòng cung cấp mã thiết bị"
	case AUTH_MISSING_IP_ADDRESS:
		return "Vui lòng cung cấp địa chỉ IP"
	case AUTH_MISSING_DEVICE_PUSH_TOKEN:
		return "Vui lòng cung cấp mã thông báo đẩy"
	case AUTH_INVALID_TOKEN:
		return "Token không hợp lệ hoặc đã hết hạn"
	case AUTH_LOGIN_FAILED:
		return "Đăng nhập thất bại"
	case AUTH_LOGOUT_FAILED:
		return "Đăng xuất thất bại"

	// Profile
	case PROFILE_NOT_FOUND:
		return "Không tìm thấy hồ sơ"
	case PROFILE_MISSING_NAME:
		return "Vui lòng nhập tên của con"
	case PROFILE_MISSING_USER_ID:
		return "Mã người dùng là bắt buộc"
	case PROFILE_MISSING_GRADE_ID:
		return "Mã lớp là bắt buộc"
	case PROFILE_MISSING_SEMESTER_ID:
		return "Mã học kỳ là bắt buộc"
	case PROFILE_INVALID_DOB:
		return "Ngày sinh không hợp lệ"
	case PROFILE_AVATAR_INVALID_FILE:
		return "Tệp ảnh đại diện không hợp lệ"
	case PROFILE_AVATAR_UPLOAD_FAILED:
		return "Tải ảnh đại diện thất bại"
	case PROFILE_MISSING_PROGRAM_ID:
		return "Mã chương trình học là bắt buộc"
	case PROFILE_INVALID_LANGUAGE:
		return "Ngôn ngữ không hợp lệ"

	// Program
	case PROGRAM_NOT_FOUND:
		return "Không tìm thấy chương trình học"
	case PROGRAM_INVALID_LANGUAGE:
		return "Ngôn ngữ chương trình học không hợp lệ"

	// Grade
	case GRADE_NOT_FOUND:
		return "Không tìm thấy lớp"
	case GRADE_INVALID_LANGUAGE:
		return "Ngôn ngữ lớp không hợp lệ"

	// Semester
	case SEMESTER_NOT_FOUND:
		return "Không tìm thấy học kỳ"
	case SEMESTER_INVALID_LANGUAGE:
		return "Ngôn ngữ học kỳ không hợp lệ"

	// Device
	case DEVICE_NOT_FOUND:
		return "Không tìm thấy thiết bị"
	case DEVICE_MISSING_UUID:
		return "Vui lòng cung cấp mã thiết bị"
	case DEVICE_MISSING_NAME:
		return "Vui lòng cung cấp tên thiết bị"
	case DEVICE_MISSING_USER_ID:
		return "Mã người dùng là bắt buộc"
	case DEVICE_NOT_OWNED:
		return "Thiết bị này không thuộc về người dùng"
	case DEVICE_ALREADY_VERIFIED:
		return "Thiết bị đã được xác minh"
	case DEVICE_REGISTRATION_FAIL:
		return "Đăng ký thiết bị thất bại"
	case DEVICE_VERIFICATION_FAIL:
		return "Xác minh thiết bị thất bại"
	case DEVICE_REVOKE_FAIL:
		return "Thu hồi thiết bị thất bại"
	case DEVICE_2FA_REQUIRED:
		return "Thiết bị này yêu cầu xác thực hai yếu tố"

	// Chapter
	case CHAPTER_NOT_FOUND:
		return "Không tìm thấy chương học"
	case CHAPTER_MISSING_ID:
		return "Mã chương là bắt buộc"
	case CHAPTER_MISSING_PROGRAM_ID:
		return "Mã chương trình là bắt buộc"
	case CHAPTER_MISSING_GRADE_ID:
		return "Mã lớp là bắt buộc"
	case CHAPTER_MISSING_LABEL:
		return "Tên chương là bắt buộc"
	case CHAPTER_MISSING_DESCRIPTION:
		return "Mô tả chương là bắt buộc"
	case CHAPTER_INVALID_DISPLAY_ORDER:
		return "Thứ tự hiển thị không hợp lệ"
	case CHAPTER_INVALID_LANGUAGE:
		return "Ngôn ngữ chương không hợp lệ"
	case CHAPTER_INVALID_TRANSLATION:
		return "Dữ liệu bản dịch chương không hợp lệ"
	case CHAPTER_TRANSLATION_ALREADY_EXISTS:
		return "Đã tồn tại bản dịch của chương cho ngôn ngữ này"
	case CHAPTER_TRANSLATION_NOT_FOUND:
		return "Không tìm thấy bản dịch chương"

	// OTP
	case OTP_NOT_FOUND:
		return "Không tìm thấy mã OTP"
	case OTP_MISSING_TYPE:
		return "Vui lòng cung cấp loại OTP"
	case OTP_MISSING_IDENTIFIER:
		return "Vui lòng nhập số điện thoại hoặc email"
	case OTP_MISSING_CODE:
		return "Vui lòng nhập mã OTP"
	case OTP_INVALID_TYPE:
		return "Loại OTP không hợp lệ"
	case OTP_INVALID_CODE:
		return "Mã OTP không chính xác"
	case OTP_EXPIRED:
		return "Mã OTP đã hết hạn"
	case OTP_ALREADY_VERIFIED:
		return "Mã OTP đã được xác thực"
	case OTP_REVOKED:
		return "Mã OTP đã bị thu hồi"
	case OTP_TOO_MANY_ATTEMPTS:
		return "Bạn đã nhập sai quá nhiều lần; vui lòng yêu cầu mã OTP mới"
	case OTP_TOO_FREQUENT:
		return "Vui lòng đợi trước khi yêu cầu mã OTP mới"
	case OTP_RATE_LIMITED:
		return "Đã vượt quá số lần yêu cầu OTP; vui lòng thử lại sau"
	case OTP_DELIVERY_FAILED:
		return "Gửi mã OTP thất bại"
	case OTP_NO_DELIVERY_CHANNEL:
		return "Không có kênh gửi phù hợp cho thông tin này"
	case OTP_GENERATION_FAILED:
		return "Không thể tạo mã OTP"

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

	// Bot (AI / chat)
	case BOT_CONNECT_FAILED:
		return "Kết nối nhà cung cấp AI thất bại"
	case BOT_CONFIG_INVALID:
		return "Cấu hình AI không hợp lệ"
	case BOT_OP_FAILED:
		return "Thao tác AI thất bại"
	case BOT_SERIALIZE_FAILED:
		return "Không thể giải mã phản hồi AI"
	case BOT_INVALID_PROMPT:
		return "Nội dung yêu cầu AI không hợp lệ"
	case BOT_CONTEXT_TOO_LARGE:
		return "Vượt quá giới hạn ngữ cảnh của mô hình AI"
	case BOT_RATE_LIMITED:
		return "Đã vượt quá giới hạn tần suất gọi AI"
	case BOT_UNSUPPORTED_OP:
		return "Thao tác AI không được nhà cung cấp hiện tại hỗ trợ"

	// Quiz
	case QUIZ_NOT_FOUND:
		return "Không tìm thấy bài kiểm tra"
	case QUIZ_MISSING_PROFILE_ID:
		return "Mã hồ sơ là bắt buộc"
	case QUIZ_MISSING_TYPE:
		return "Vui lòng chọn loại bài kiểm tra"
	case QUIZ_INVALID_TYPE:
		return "Loại bài kiểm tra không hợp lệ"
	case QUIZ_INVALID_LANGUAGE:
		return "Ngôn ngữ bài kiểm tra không hợp lệ"
	case QUIZ_MISSING_ANSWERS:
		return "Vui lòng cung cấp câu trả lời"
	case QUIZ_INVALID_ANSWERS:
		return "Câu trả lời không hợp lệ"
	case QUIZ_ALREADY_SUBMITTED:
		return "Bài kiểm tra này đã được nộp"
	case QUIZ_NOT_OWNED:
		return "Bài kiểm tra không thuộc về hồ sơ này"
	case QUIZ_PREVIOUS_NOT_FOUND:
		return "Không tìm thấy bài kiểm tra trước đó"
	case QUIZ_PREVIOUS_NOT_GRADED:
		return "Bài kiểm tra trước đó chưa được chấm điểm"
	case QUIZ_GENERATION_FAILED:
		return "Tạo câu hỏi bài kiểm tra thất bại"
	case QUIZ_GRADING_FAILED:
		return "Chấm điểm bài kiểm tra thất bại"
	case QUIZ_PROFILE_NOT_CONFIGURED:
		return "Hồ sơ cần có chương trình học, lớp và học kỳ trước khi tạo bài kiểm tra"
	case QUIZ_INVALID_TYPE_OF_QUIZ:
		return "Hình thức bài học không hợp lệ"

	// Job
	case JOB_NOT_FOUND:
		return "Không tìm thấy job"
	case JOB_ALREADY_PAUSED:
		return "Job đã ở trạng thái tạm dừng"
	case JOB_NOT_PAUSED:
		return "Job không ở trạng thái tạm dừng"
	case JOB_IN_FLIGHT:
		return "Job đang chạy một lần thực thi khác"
	case JOB_RUNTIME_UNAVAILABLE:
		return "Hệ thống job chưa khởi động hoặc đang dừng"
	case JOB_MISSING_NAME:
		return "Tên job là bắt buộc"
	case JOB_TRIGGER_FAILED:
		return "Kích hoạt job thất bại"
	case TASK_HANDLER_NOT_FOUND:
		return "Không tìm thấy task handler được đăng ký"
	case TASK_QUEUE_FULL:
		return "Hàng đợi task đã đầy; vui lòng thử lại sau"
	case TASK_MISSING_NAME:
		return "Tên task là bắt buộc"
	case TASK_ENQUEUE_FAILED:
		return "Đưa task vào hàng đợi thất bại"
	default:
		return ""
	}
}
