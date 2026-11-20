package questionpool

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (14400–14499). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Question pool (admin bank) — 14400-14499 (ma_question_pools).
	// 14400 / 14401 were QUESTION_POOL_MISSING_EXAM_ID / QUESTION_POOL_INVALID_SHAPE
	// (pool copy, never returned by a route); reused 2026-10-26.
	// debug names every offending question by path (questions[3].…).
	// QuestionPoolEmptyList: a create sent no questions.
	QuestionPoolEmptyList status.StatusCode = 14400
	// QuestionPoolInvalid: a question breaks a general rule (text, grade,
	// level, lengths, tags, topic).
	QuestionPoolInvalid status.StatusCode = 14401
	// QuestionPoolTooMany: a create sent more questions than one call takes.
	QuestionPoolTooMany status.StatusCode = 14402
	// QuestionPoolInvalidKind: an unknown question_kind (create or list filter).
	QuestionPoolInvalidKind status.StatusCode = 14403
	// QuestionPoolInvalidOptions: options missing / present against the kind,
	// too few or too many, a blank or repeated id, an empty option.
	QuestionPoolInvalidOptions status.StatusCode = 14404
	// QuestionPoolInvalidAnswer: correct_answer has the wrong shape for the
	// kind, names an option that does not exist, or the wrong count.
	QuestionPoolInvalidAnswer status.StatusCode = 14405
	// QuestionPoolInvalidImageKey: an image key outside question-images/ or
	// shaped like a URL.
	QuestionPoolInvalidImageKey status.StatusCode = 14406
	// QuestionPoolDuplicate: the same question twice in one request, or one the
	// bank already holds live.
	QuestionPoolDuplicate status.StatusCode = 14407
	// QuestionPoolInvalidFilter: a list filter out of range (grade, status,
	// source).
	QuestionPoolInvalidFilter status.StatusCode = 14408
	// QuestionPoolImageMissing: /questions/images/upload carried no "image" file part.
	QuestionPoolImageMissing status.StatusCode = 14409
	// QuestionPoolImageInvalidFile: not a JPEG / PNG / GIF / WebP image (by
	// extension, declared type and content), or over the size limit.
	QuestionPoolImageInvalidFile status.StatusCode = 14410
	// QuestionPoolImageUploadFailed: storage refused the upload.
	QuestionPoolImageUploadFailed status.StatusCode = 14411
	// QuestionPoolMissingID: an update named no question_id.
	QuestionPoolMissingID status.StatusCode = 14412
	// QuestionPoolNotFound: no live question has that question_id.
	QuestionPoolNotFound status.StatusCode = 14413
	// QuestionPoolNothingToUpdate: an update named no field to change.
	QuestionPoolNothingToUpdate status.StatusCode = 14414
	// QuestionPoolInvalidStatus: mark-status named no question_status, or an unknown one.
	QuestionPoolInvalidStatus status.StatusCode = 14415
	// QuestionPoolImportFileMissing: validate-file / import-file got no "file" part.
	QuestionPoolImportFileMissing status.StatusCode = 14416
	// QuestionPoolImportFileInvalid: the file is not .json, cannot be read,
	// is not valid JSON, or is not an array of objects.
	QuestionPoolImportFileInvalid status.StatusCode = 14417
	// QuestionPoolImportFileTooLarge: the file is over the size limit.
	QuestionPoolImportFileTooLarge status.StatusCode = 14418
	// QuestionPoolImportInvalid: import-file found at least one invalid
	// question, so nothing was imported.
	QuestionPoolImportInvalid status.StatusCode = 14419
	// QuestionPoolDeleteFilterRequired: a delete by filter named no
	// filter; refused so one call cannot wipe the whole bank.
	QuestionPoolDeleteFilterRequired status.StatusCode = 14420
	// QuestionPoolImageInUse: images/delete named a key a live question
	// still shows (stem or option image).
	QuestionPoolImageInUse status.StatusCode = 14421
	// QuestionPoolImageDeleteFailed: storage refused the delete.
	QuestionPoolImageDeleteFailed status.StatusCode = 14422
	// QuestionPoolInvalidTranslation: question_language or a translation
	// breaks a rule (unknown language, the question's own language, missing
	// question_text, options that are not exactly the question's ids).
	QuestionPoolInvalidTranslation status.StatusCode = 14423
)

func init() {
	status.Register(QuestionPoolEmptyList, "At least one question is required", "Cần ít nhất một câu hỏi")
	status.Register(QuestionPoolInvalid, "A question is invalid", "Câu hỏi không hợp lệ")
	status.Register(QuestionPoolTooMany, "Too many questions in one request", "Quá nhiều câu hỏi trong một lần gửi")
	status.Register(QuestionPoolInvalidKind, "The question kind is invalid", "Dạng câu hỏi không hợp lệ")
	status.Register(QuestionPoolInvalidOptions, "The question's options are invalid", "Các lựa chọn của câu hỏi không hợp lệ")
	status.Register(QuestionPoolInvalidAnswer, "The question's correct answer is invalid", "Đáp án đúng của câu hỏi không hợp lệ")
	status.Register(QuestionPoolInvalidImageKey, "The image key is invalid", "Khóa ảnh không hợp lệ")
	status.Register(QuestionPoolDuplicate, "The question already exists in the question bank", "Câu hỏi đã có trong ngân hàng câu hỏi")
	status.Register(QuestionPoolInvalidFilter, "The question filter is invalid", "Bộ lọc câu hỏi không hợp lệ")
	status.Register(QuestionPoolImageMissing, "An image file is required", "Cần gửi kèm tệp ảnh")
	status.Register(QuestionPoolImageInvalidFile, "The image must be a JPEG, PNG, GIF or WebP file of at most 5 MB", "Ảnh phải là tệp JPEG, PNG, GIF hoặc WebP, tối đa 5 MB")
	status.Register(QuestionPoolImageUploadFailed, "The image could not be uploaded", "Không tải ảnh lên được")
	status.Register(QuestionPoolMissingID, "question_pool_id is required", "Thiếu question_pool_id")
	status.Register(QuestionPoolNotFound, "The question does not exist", "Câu hỏi không tồn tại")
	status.Register(QuestionPoolNothingToUpdate, "Nothing to update", "Không có thông tin nào để cập nhật")
	status.Register(QuestionPoolInvalidStatus, "question_status must be PENDING_REVIEW, APPROVED or REJECTED", "question_status phải là PENDING_REVIEW, APPROVED hoặc REJECTED")
	status.Register(QuestionPoolImportFileMissing, "Please attach a .json file", "Vui lòng đính kèm file .json")
	status.Register(QuestionPoolImportFileInvalid, "The file must be a .json file holding an array of questions", "File phải là file .json chứa một mảng câu hỏi")
	status.Register(QuestionPoolImportFileTooLarge, "The file is too large", "File quá lớn")
	status.Register(QuestionPoolImportInvalid, "The file has invalid questions; nothing was imported", "File có câu hỏi không hợp lệ, chưa có câu nào được import")
	status.Register(QuestionPoolDeleteFilterRequired, "Please choose at least one filter for the questions to delete", "Vui lòng chọn ít nhất một bộ lọc cho các câu hỏi cần xoá")
	status.Register(QuestionPoolImageInUse, "The image is still used by a question", "Ảnh đang được câu hỏi sử dụng, không thể xoá")
	status.Register(QuestionPoolImageDeleteFailed, "Could not delete the image", "Không xoá được ảnh")
	status.Register(QuestionPoolInvalidTranslation, "The question language or a translation is invalid", "Ngôn ngữ hoặc bản dịch của câu hỏi không hợp lệ")
}
