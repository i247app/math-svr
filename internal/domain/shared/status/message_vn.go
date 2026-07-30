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
	case USER_AVATAR_CONFLICT:
		return "Chỉ chọn một trong hai: tệp ảnh hoặc đường dẫn ảnh"
	case USER_AVATAR_INVALID_REFERENCE:
		return "Đường dẫn ảnh đại diện không hợp lệ"
	case USER_INVALID_ROLE:
		return "Vai trò không hợp lệ, các vai trò được hỗ trợ: {roles}"
	case USER_EMAIL_NOT_VERIFIED:
		return "Email chưa được xác thực. Vui lòng xác thực email bằng OTP trước khi đăng ký với email này"

	// Auth
	case AUTH_MISSING_LOGIN_NAME:
		return "Vui lòng nhập tên đăng nhập"
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
	case PROFILE_AVATAR_CONFLICT:
		return "Chỉ chọn một trong hai: tệp ảnh hoặc đường dẫn ảnh"
	case PROFILE_AVATAR_INVALID_REFERENCE:
		return "Đường dẫn ảnh đại diện không hợp lệ"
	case PROFILE_MISSING_PROGRAM_ID:
		return "Mã chương trình học là bắt buộc"
	case PROFILE_INVALID_LANGUAGE:
		return "Ngôn ngữ không hợp lệ"
	case PROFILE_CODE_TAKEN:
		return "Mã hồ sơ đã được sử dụng"
	case PROFILE_CODE_GENERATION_FAILED:
		return "Không thể tạo mã hồ sơ duy nhất"

	// Program
	case PROGRAM_NOT_FOUND:
		return "Không tìm thấy chương trình học"
	case PROGRAM_MISSING_ID:
		return "Mã chương trình học là bắt buộc"
	case PROGRAM_MISSING_LABEL:
		return "Tên chương trình học là bắt buộc"
	case PROGRAM_MISSING_DESCRIPTION:
		return "Mô tả chương trình học là bắt buộc"
	case PROGRAM_INVALID_DISPLAY_ORDER:
		return "Thứ tự hiển thị của chương trình học không hợp lệ"

	// Grade
	case GRADE_NOT_FOUND:
		return "Không tìm thấy lớp"
	case GRADE_MISSING_ID:
		return "Mã lớp là bắt buộc"
	case GRADE_MISSING_LABEL:
		return "Tên lớp là bắt buộc"
	case GRADE_MISSING_DESCRIPTION:
		return "Mô tả lớp là bắt buộc"
	case GRADE_INVALID_DISPLAY_ORDER:
		return "Thứ tự hiển thị của lớp không hợp lệ"

	// Semester
	case SEMESTER_NOT_FOUND:
		return "Không tìm thấy học kỳ"
	case SEMESTER_MISSING_ID:
		return "Mã học kỳ là bắt buộc"
	case SEMESTER_MISSING_NAME:
		return "Tên học kỳ là bắt buộc"
	case SEMESTER_MISSING_DESCRIPTION:
		return "Mô tả học kỳ là bắt buộc"
	case SEMESTER_INVALID_DISPLAY_ORDER:
		return "Thứ tự hiển thị của học kỳ không hợp lệ"

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
	case DEVICE_NOT_TRUSTED:
		return "Thiết bị được chọn chưa được tin cậy (trusted)"

	// School
	case SCHOOL_NOT_FOUND:
		return "Không tìm thấy trường học"
	case SCHOOL_MISSING_ID:
		return "Mã trường học là bắt buộc"
	case SCHOOL_MISSING_NAME:
		return "Tên trường học là bắt buộc"
	case SCHOOL_NAME_TOO_LONG:
		return "Tên trường học quá dài"
	case SCHOOL_INVALID_STATUS:
		return "Trạng thái trường học không hợp lệ"
	case SCHOOL_DESCRIPTION_TOO_LONG:
		return "Mô tả trường học quá dài"
	case SCHOOL_PROFILE_LINK_FAILED:
		return "Liên kết hồ sơ với trường học thất bại"

	// Chapter
	case CHAPTER_NOT_FOUND:
		return "Không tìm thấy chương học"
	case CHAPTER_MISSING_ID:
		return "Mã chương là bắt buộc"
	case CHAPTER_MISSING_PROGRAM_ID:
		return "Mã chương trình là bắt buộc"
	case CHAPTER_MISSING_GRADE_ID:
		return "Mã lớp là bắt buộc"
	case CHAPTER_MISSING_SEMESTER_ID:
		return "Mã học kỳ là bắt buộc"
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
	case OTP_DISABLED:
		return "Xác thực 2 bước đã bị tắt"
	case OTP_TARGET_DEVICE_REQUIRES_LOGIN2FA:
		return "target_device_id chỉ được hỗ trợ với type LOGIN_2FA"

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
	case NOTIFICATION_SEND_FAILED:
		return "Gửi thông báo thất bại"
	case NOTIFICATION_NO_DEVICE_TOKEN:
		return "Người nhận không có thiết bị nào để nhận thông báo"
	case NOTIFICATION_DISABLED:
		return "Dịch vụ thông báo đang bị tắt"
	case NOTIFICATION_NOT_OWNED:
		return "Thông báo không thuộc về người dùng hiện tại"
	case NOTIFICATION_CONFIG_INVALID:
		return "Cấu hình thông báo không hợp lệ"
	case NOTIFICATION_CONNECT_FAILED:
		return "Không thể kết nối tới dịch vụ thông báo"

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
	case QUIZ_ANSWER_SCHEMA_INVALID:
		return "Định dạng câu trả lời không hợp lệ"
	case QUIZ_ANALYTICS_MISSING_PROFILE:
		return "Hồ sơ học sinh là bắt buộc."
	case QUIZ_ANALYTICS_PROFILE_NOT_OWNED:
		return "Hồ sơ này không thuộc về bạn."
	case QUIZ_ANALYTICS_INVALID_DATE_RANGE:
		return "Khoảng thời gian không hợp lệ."
	case QUIZ_ANALYTICS_INVALID_TZ:
		return "Múi giờ không hợp lệ."
	case QUIZ_ANALYTICS_INVALID_PURPOSE:
		return "Loại quiz không hợp lệ."

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

	// Sequence
	case SEQ_NOT_FOUND:
		return "Không tìm thấy sequence"
	case SEQ_MISSING_NAME:
		return "Tên sequence là bắt buộc"
	case SEQ_GENERATION_FAILED:
		return "Sinh id mới thất bại"

	// Classroom
	case CLASSROOM_NOT_FOUND:
		return "Không tìm thấy lớp học"
	case CLASSROOM_MISSING_ID:
		return "Mã lớp học là bắt buộc"
	case CLASSROOM_MISSING_NAME:
		return "Tên lớp học là bắt buộc"
	case CLASSROOM_NAME_TOO_LONG:
		return "Tên lớp học quá dài"
	case CLASSROOM_DESCRIPTION_TOO_LONG:
		return "Mô tả lớp học quá dài"
	case CLASSROOM_MISSING_OWNER_PROFILE_ID:
		return "Hồ sơ chủ lớp là bắt buộc"
	case CLASSROOM_INVALID_OWNER_ROLE:
		return "Chỉ hồ sơ giáo viên mới có thể làm chủ lớp học"
	case CLASSROOM_INVALID_PROGRAM:
		return "Chương trình không hợp lệ"
	case CLASSROOM_INVALID_GRADE:
		return "Khối lớp không hợp lệ"
	case CLASSROOM_ALREADY_ARCHIVED:
		return "Lớp học đã được lưu trữ"
	case CLASSROOM_NOT_ARCHIVED:
		return "Lớp học chưa được lưu trữ"
	case CLASSROOM_ALREADY_DELETED:
		return "Lớp học đã bị xoá"
	case CLASSROOM_MAX_MEMBERS_REACHED:
		return "Lớp học đã đạt giới hạn thành viên"
	case CLASSROOM_CODE_INVALID:
		return "Mã mời không hợp lệ"
	case CLASSROOM_CODE_EXPIRED:
		return "Mã mời đã hết hạn"
	case CLASSROOM_CODE_DISABLED:
		return "Mã mời đã bị vô hiệu hoá"
	case CLASSROOM_CODE_GENERATION_FAILED:
		return "Tạo mã mời thất bại"
	case CLASSROOM_PERMISSION_DENIED:
		return "Bạn không có quyền thực hiện thao tác này trên lớp học"
	case CLASSROOM_OWNER_CANNOT_LEAVE:
		return "Chủ lớp phải chuyển quyền sở hữu trước khi rời lớp"
	case CLASSROOM_OWNER_TRANSFER_TO_NON_MEMBER:
		return "Chỉ có thể chuyển quyền sở hữu cho thành viên đang hoạt động"
	case CLASSROOM_INVALID_MAX_MEMBERS:
		return "Giới hạn thành viên phải là số dương"
	case CLASSROOM_INVALID_SCHOOL:
		return "Không tìm thấy trường học"
	case CLASSROOM_CODE_TAKEN:
		return "Mã mời đã được sử dụng"
	case CLASSROOM_PROGRAM_DUPLICATE:
		return "Danh sách chương trình bị trùng"
	case CLASSROOM_PROGRESS_INVALID_DATE_RANGE:
		return "Khoảng thời gian không hợp lệ."
	case CLASSROOM_PROGRESS_INVALID_TZ:
		return "Múi giờ không hợp lệ."
	case CLASSROOM_PROGRESS_INVALID_PURPOSE:
		return "Loại bài không hợp lệ."

	// Classroom member
	case CLASSROOM_MEMBER_NOT_FOUND:
		return "Không tìm thấy thành viên lớp học"
	case CLASSROOM_MEMBER_MISSING_ID:
		return "Mã thành viên là bắt buộc"
	case CLASSROOM_MEMBER_MISSING_CLASSROOM_ID:
		return "Mã lớp học là bắt buộc"
	case CLASSROOM_MEMBER_MISSING_PROFILE_ID:
		return "Mã hồ sơ là bắt buộc"
	case CLASSROOM_MEMBER_ALREADY_MEMBER:
		return "Hồ sơ đã là thành viên của lớp học"
	case CLASSROOM_MEMBER_NOT_MEMBER:
		return "Hồ sơ không phải là thành viên của lớp học"
	case CLASSROOM_MEMBER_INVALID_ROLE:
		return "Vai trò thành viên không hợp lệ"
	case CLASSROOM_MEMBER_CANNOT_REMOVE_OWNER:
		return "Không thể xoá chủ lớp"
	case CLASSROOM_MEMBER_CANNOT_DEMOTE_OWNER:
		return "Không thể đổi vai trò chủ lớp; hãy chuyển quyền sở hữu trước"

	// Classroom invitation
	case CLASSROOM_INVITATION_NOT_FOUND:
		return "Không tìm thấy lời mời"
	case CLASSROOM_INVITATION_MISSING_ID:
		return "Mã lời mời là bắt buộc"
	case CLASSROOM_INVITATION_MISSING_TARGET:
		return "Lời mời phải có hồ sơ hoặc định danh người nhận"
	case CLASSROOM_INVITATION_INVALID_IDENTIFIER:
		return "Định danh người nhận không hợp lệ"
	case CLASSROOM_INVITATION_INVALID_IDENTIFIER_TYPE:
		return "Loại định danh người nhận không hợp lệ"
	case CLASSROOM_INVITATION_INVALID_ROLE:
		return "Vai trò đề xuất không hợp lệ"
	case CLASSROOM_INVITATION_ALREADY_RESPONDED:
		return "Lời mời đã được phản hồi"
	case CLASSROOM_INVITATION_EXPIRED:
		return "Lời mời đã hết hạn"
	case CLASSROOM_INVITATION_NOT_PENDING:
		return "Lời mời không ở trạng thái chờ"
	case CLASSROOM_INVITATION_NOT_INVITEE:
		return "Bạn không phải người được mời"
	case CLASSROOM_INVITATION_ALREADY_INVITED:
		return "Đã tồn tại lời mời đang chờ cho người này"
	case CLASSROOM_INVITATION_TOKEN_INVALID:
		return "Token lời mời không hợp lệ"
	case CLASSROOM_INVITATION_TOKEN_GENERATION_FAILED:
		return "Tạo token lời mời thất bại"
	case CLASSROOM_INVITATION_PERMISSION_DENIED:
		return "Bạn không có quyền thao tác trên lời mời này"

	// Classroom join request
	case CLASSROOM_JOIN_REQUEST_NOT_FOUND:
		return "Không tìm thấy yêu cầu tham gia"
	case CLASSROOM_JOIN_REQUEST_NOT_PENDING:
		return "Yêu cầu tham gia không ở trạng thái chờ duyệt"
	case CLASSROOM_JOIN_REQUEST_ALREADY_PENDING:
		return "Đã tồn tại yêu cầu tham gia đang chờ duyệt cho lớp học này"
	case CLASSROOM_JOIN_REQUEST_PERMISSION_DENIED:
		return "Bạn không có quyền thao tác trên yêu cầu tham gia này"

	// Classroom exercise
	case CLASSROOM_EXERCISE_NOT_FOUND:
		return "Không tìm thấy bài tập lớp học"
	case CLASSROOM_EXERCISE_MISSING_ID:
		return "Mã bài tập là bắt buộc"
	case CLASSROOM_EXERCISE_MISSING_CLASSROOM_ID:
		return "Mã lớp học là bắt buộc"
	case CLASSROOM_EXERCISE_MISSING_TITLE:
		return "Tên bài tập là bắt buộc"
	case CLASSROOM_EXERCISE_TITLE_TOO_LONG:
		return "Tên bài tập quá dài"
	case CLASSROOM_EXERCISE_MISSING_CHAPTER_NAME:
		return "Tên chương là bắt buộc"
	case CLASSROOM_EXERCISE_CHAPTER_NAME_TOO_LONG:
		return "Tên chương quá dài"
	case CLASSROOM_EXERCISE_MISSING_LESSON_NAME:
		return "Tên bài học là bắt buộc"
	case CLASSROOM_EXERCISE_LESSON_NAME_TOO_LONG:
		return "Tên bài học quá dài"
	case CLASSROOM_EXERCISE_INVALID_NUM_QUESTIONS:
		return "Số lượng câu hỏi không hợp lệ"
	case CLASSROOM_EXERCISE_INVALID_DATE_RANGE:
		return "Ngày kết thúc phải sau ngày bắt đầu"
	case CLASSROOM_EXERCISE_NOTE_TOO_LONG:
		return "Ghi chú quá dài"
	case CLASSROOM_EXERCISE_INVALID_PROGRAM:
		return "Bộ sách không hợp lệ"
	case CLASSROOM_EXERCISE_PROGRAM_NOT_IN_CLASSROOM:
		return "Bộ sách không thuộc lớp học này"
	case CLASSROOM_EXERCISE_ALREADY_DELETED:
		return "Bài tập đã bị xoá"
	case CLASSROOM_EXERCISE_PERMISSION_DENIED:
		return "Bạn không có quyền thao tác trên bài tập này"
	case CLASSROOM_EXERCISE_GENERATION_FAILED:
		return "Tạo bài tập tự động thất bại"
	case CLASSROOM_EXERCISE_INVALID_VISIBILITY:
		return "Chế độ hiển thị phải là PUBLIC hoặc PRIVATE"
	case CLASSROOM_EXERCISE_PRIVATE_DENIED:
		return "Chỉ người tạo mới có quyền truy cập bài tập riêng tư này"
	case CLASSROOM_EXERCISE_INVALID_PURPOSE:
		return "Mục đích phải là HOMEWORK hoặc EXAM"
	case CLASSROOM_EXERCISE_SUBMISSION_NOT_FOUND:
		return "Không tìm thấy bài nộp"
	case CLASSROOM_EXERCISE_SUBMISSION_MISSING_ID:
		return "classroom_exercise_submission_id là bắt buộc"
	case CLASSROOM_EXERCISE_SUBMISSION_MISSING_EXERCISE_ID:
		return "classroom_exercise_id là bắt buộc"
	case CLASSROOM_EXERCISE_SUBMISSION_MISSING_ANSWERS:
		return "answers là bắt buộc"
	case CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS:
		return "Dữ liệu câu trả lời không hợp lệ"
	case CLASSROOM_EXERCISE_SUBMISSION_ALREADY_EXISTS:
		return "Bạn đã nộp bài tập này rồi"
	case CLASSROOM_EXERCISE_SUBMISSION_WINDOW_NOT_OPEN:
		return "Bài tập chưa mở để làm"
	case CLASSROOM_EXERCISE_SUBMISSION_WINDOW_CLOSED:
		return "Bài tập đã hết hạn nộp"
	case CLASSROOM_EXERCISE_SUBMISSION_EXERCISE_UNAVAILABLE:
		return "Bài tập hiện không khả dụng để nộp"
	case CLASSROOM_EXERCISE_SUBMISSION_PERMISSION_DENIED:
		return "Bạn không có quyền truy cập bài nộp này"
	case CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED:
		return "Chấm bài tự động thất bại"
	case CLASSROOM_EXERCISE_SUBMISSION_NOTE_TOO_LONG:
		return "Ghi chú quá dài"
	case CLASSROOM_EXERCISE_SUBMISSION_ALREADY_DELETED:
		return "Bài nộp đã bị xoá"

	// Server lifecycle
	case SERVER_SHUTTING_DOWN:
		return "Máy chủ đang tắt"

	// Home layout
	case HOME_MISSING_PROFILE_ID:
		return "Thiếu profile_id"
	case HOME_PROFILE_NOT_FOUND:
		return "Không tìm thấy hồ sơ"
	case HOME_PROFILE_NOT_OWNED:
		return "Hồ sơ này không thuộc về người dùng hiện tại"
	case HOME_UNSUPPORTED_ROLE:
		return "Vai trò hồ sơ không được hỗ trợ cho trang chủ"

	// Socket / WebSocket
	case SOCKET_UNAUTHORIZED:
		return "Kết nối WebSocket không được xác thực"
	case SOCKET_MISSING_TOPIC:
		return "Thiếu topic"
	case SOCKET_TOPIC_FORBIDDEN:
		return "Không được phép đăng ký topic này"
	case SOCKET_INVALID_MESSAGE:
		return "Tin nhắn WebSocket không hợp lệ"
	case SOCKET_TOO_MANY_CONNECTIONS:
		return "Quá nhiều kết nối đang hoạt động"
	case SOCKET_DISABLED:
		return "Kênh realtime đang bị tắt"
	default:
		return ""
	}
}
