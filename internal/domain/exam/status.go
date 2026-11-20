package exam

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (13700–13799). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Exam — 13700-13799.
	// 13600 was already taken by PRESENCE_*, so the exam aggregate claims
	// the next free 100-block rather than the 13600 the planning notes
	// originally reserved.
	ExamNotFound         status.StatusCode = 13700
	ExamAttemptNotFound  status.StatusCode = 13701
	ExamAttemptNotOwned  status.StatusCode = 13702
	ExamAlreadySubmitted status.StatusCode = 13703
	ExamMissingAttemptID status.StatusCode = 13704
	ExamMissingAnswers   status.StatusCode = 13705
	ExamInvalidAnswers   status.StatusCode = 13706
	ExamMissingExamType  status.StatusCode = 13707
	ExamInvalidExamType  status.StatusCode = 13708
	ExamInvalidGrade     status.StatusCode = 13709
	// 13710 was EXAM_INVALID_LEVEL. Retired with the level axis (no rule
	// defines it yet); left unassigned so an old client log line keeps its
	// meaning. Reuse the number only when level itself comes back.
	ExamInvalidNumQuestions   status.StatusCode = 13711
	ExamGenerationFailed      status.StatusCode = 13712
	ExamGradingFailed         status.StatusCode = 13713
	ExamNoQuestions           status.StatusCode = 13714
	ExamProfileNotFound       status.StatusCode = 13715
	ExamProfileNotOwned       status.StatusCode = 13716
	ExamMissingProfileID      status.StatusCode = 13717
	ExamStatsNotFound         status.StatusCode = 13718
	ExamAnalyticsInvalidRange status.StatusCode = 13719
	ExamAnalyticsInvalidTz    status.StatusCode = 13720
	ExamJourneyNotFound       status.StatusCode = 13721
	ExamJourneyNotOwned       status.StatusCode = 13722
	ExamJourneyAlreadyEnded   status.StatusCode = 13723
	ExamMissingJourneyID      status.StatusCode = 13724
	ExamInvalidJourneyStatus  status.StatusCode = 13725
	ExamAmbiguousDetailID     status.StatusCode = 13726
	ExamPracticeNoBase        status.StatusCode = 13727
	ExamJourneyAlreadyActive  status.StatusCode = 13728
	// 13729 was EXAM_ATTEMPT_IN_PROGRESS ("one open sitting per journey"),
	// which REFUSED a hand-out while one was open. Retired; the rule is
	// back as a non-error: generate hands the open sitting back instead
	// (GenerateExamRes.resumed). Do not reuse the number.
	ExamJourneyNotComplete status.StatusCode = 13730
	// ExamInvalidLevel lives at 13731, not at the retired 13710: the
	// axis came back as a client-stated value, not the derived one 13710
	// once described.
	ExamInvalidLevel status.StatusCode = 13731
	// ExamGuestProfileNotOwned: an unauthenticated caller named a
	// profile_id. A guest may only ever be handed their own, freshly
	// opened profile; naming someone else's is how a stranger would read
	// another child's exams.
	ExamGuestProfileNotOwned status.StatusCode = 13732
	// ExamGuestTypeNotAllowed: a guest asked for a round that is not
	// ASSESSMENT. The guest surface is a trial of the product, not the
	// product — practice and grade reviews belong to a registered child
	// whose progress is being tracked.
	ExamGuestTypeNotAllowed status.StatusCode = 13733
	// ExamGuestDailyLimit: this guest has had as many rounds today as
	// a trial is worth. Every one of them is a paid model call made for
	// someone who has not registered, so the ceiling is the product's,
	// not a per-user setting.
	ExamGuestDailyLimit status.StatusCode = 13734
	// ExamMissingGrade: a request that is answered per grade
	// (/exams/grade/levels) named none.
	ExamMissingGrade status.StatusCode = 13735
	// ExamReviewNoAnswers: an AI review was asked for a journey with
	// no answered question yet — there is nothing to review.
	ExamReviewNoAnswers status.StatusCode = 13736
	// ExamReviewFailed: the model's review could not be used (not the
	// JSON asked for, or empty).
	ExamReviewFailed status.StatusCode = 13737
	// ExamReviewJourneyNotComplete: an AI review was asked for a
	// journey that is not COMPLETE — only a finished journey is reviewed.
	// (EXAM_JOURNEY_NOT_COMPLETE is worded for practice, so not reused.)
	ExamReviewJourneyNotComplete status.StatusCode = 13738
	// ExamReviewAlreadyExists: the journey already holds both its short
	// and long AI review — it is read from the journey, not generated again.
	ExamReviewAlreadyExists status.StatusCode = 13739
	// ExamMissingExamID: an exam pool read named no exam_id.
	ExamMissingExamID status.StatusCode = 13740
	// ExamMissingIsVerify: /exams/pools/mark-verify sent no is_verify.
	ExamMissingIsVerify status.StatusCode = 13741
	// ExamPoolInvalidQuestions: a corrected question set is empty or does
	// not keep the stored set's question numbers and answer labels.
	ExamPoolInvalidQuestions status.StatusCode = 13742
	// 13743–13746 are RETIRED (EXAM_PROMPT_* of /exams/prompts/*, replaced by
	// PROMPT_* 14300+ when the prompts left the exam aggregate). Do not reuse.
	// ExamQuestionBankInsufficient: QUESTION_POOL_ENABLED is on and the
	// bank holds too few approved questions for the paper (the debug field
	// names the group and the shortfall). The model is never called instead.
	ExamQuestionBankInsufficient status.StatusCode = 13747
	// ExamPoolBankReadonly: /exams/pools/verify on a set built from the
	// question bank — correct the question in the bank instead.
	ExamPoolBankReadonly status.StatusCode = 13748
)

func init() {
	status.Register(ExamNotFound, "Exam not found", "Không tìm thấy bài kiểm tra")
	status.Register(ExamAttemptNotFound, "Exam attempt not found", "Không tìm thấy lượt làm bài")
	status.Register(ExamAttemptNotOwned, "This exam attempt does not belong to the current profile", "Lượt làm bài này không thuộc về hồ sơ hiện tại")
	status.Register(ExamAlreadySubmitted, "This exam has already been submitted", "Bài kiểm tra này đã được nộp")
	status.Register(ExamMissingAttemptID, "elink_id is required", "Thiếu elink_id")
	status.Register(ExamMissingAnswers, "answers is required", "Thiếu câu trả lời")
	status.Register(ExamInvalidAnswers, "answers payload is invalid", "Dữ liệu câu trả lời không hợp lệ")
	status.Register(ExamMissingExamType, "exam_type is required", "Thiếu loại bài kiểm tra")
	status.Register(ExamInvalidExamType, "exam_type must be one of ASSESSMENT, PRACTICE, EXAM", "Loại bài kiểm tra không hợp lệ")
	status.Register(ExamInvalidGrade, "grade must be between 0 and 5", "Cấp lớp phải nằm trong khoảng 0 đến 5")
	status.Register(ExamInvalidNumQuestions, "num_questions is invalid", "Số lượng câu hỏi không hợp lệ")
	status.Register(ExamGenerationFailed, "Failed to generate the exam", "Tạo bài kiểm tra thất bại")
	status.Register(ExamGradingFailed, "Failed to grade the exam", "Chấm bài kiểm tra thất bại")
	status.Register(ExamNoQuestions, "This exam has no questions to grade", "Bài kiểm tra này không có câu hỏi để chấm")
	status.Register(ExamProfileNotFound, "Profile not found", "Không tìm thấy hồ sơ")
	status.Register(ExamProfileNotOwned, "This profile does not belong to the current user", "Hồ sơ này không thuộc về người dùng hiện tại")
	status.Register(ExamMissingProfileID, "profile_id is required", "Thiếu profile_id")
	status.Register(ExamStatsNotFound, "No exam statistics found for this profile", "Chưa có thống kê bài kiểm tra cho hồ sơ này")
	status.Register(ExamAnalyticsInvalidRange, "Invalid date range", "Khoảng thời gian không hợp lệ")
	status.Register(ExamAnalyticsInvalidTz, "Invalid timezone offset", "Múi giờ không hợp lệ")
	status.Register(ExamJourneyNotFound, "Exam journey not found", "Không tìm thấy hành trình kiểm tra")
	status.Register(ExamJourneyNotOwned, "This exam journey does not belong to the current profile", "Hành trình kiểm tra này không thuộc về hồ sơ hiện tại")
	status.Register(ExamJourneyAlreadyEnded, "This exam journey has already ended", "Hành trình kiểm tra này đã kết thúc")
	status.Register(ExamMissingJourneyID, "esess_id is required", "Thiếu esess_id")
	status.Register(ExamInvalidJourneyStatus, "status must be ACTIVE, COMPLETE or CANCEL", "Trạng thái phải là ACTIVE, COMPLETE hoặc CANCEL")
	status.Register(ExamAmbiguousDetailID, "Send either elink_id or esess_id, not both", "Chỉ gửi một trong elink_id hoặc esess_id")
	status.Register(ExamPracticeNoBase, "Nothing to practise yet — submit an exam in this journey first", "Chưa có bài để luyện tập — hãy nộp một bài trong hành trình này trước")
	status.Register(ExamJourneyAlreadyActive, "A journey of this type is already open — end it before reopening another", "Đang có một hành trình mở — hãy kết thúc nó trước khi mở lại hành trình khác")
	status.Register(ExamJourneyNotComplete, "Practice opens once the journey is complete", "Chỉ luyện tập được khi hành trình đã hoàn thành")
	status.Register(ExamInvalidLevel, "level must be between 0 and 9", "Cấp độ phải từ 1 đến 10")
	status.Register(ExamGuestProfileNotOwned, "Please sign in to open this profile's exams", "Vui lòng đăng nhập để xem bài của hồ sơ này")
	status.Register(ExamGuestTypeNotAllowed, "Sign up to unlock practice and grade reviews", "Hãy đăng ký để mở bài luyện tập và bài kiểm tra theo lớp")
	status.Register(ExamGuestDailyLimit, "You have used all of today's free exams. Sign up to keep going", "Bé đã dùng hết số bài miễn phí hôm nay. Hãy đăng ký để tiếp tục")
	status.Register(ExamMissingGrade, "Grade is required", "Vui lòng chọn lớp")
	status.Register(ExamReviewNoAnswers, "There are no answers to review yet. Finish an exam first", "Chưa có câu trả lời nào để đánh giá. Hãy hoàn thành một bài kiểm tra trước")
	status.Register(ExamReviewFailed, "Could not create the review. Please try again", "Không tạo được đánh giá. Vui lòng thử lại")
	status.Register(ExamReviewJourneyNotComplete, "A review is available once the journey is complete", "Chỉ xem được đánh giá khi hành trình đã hoàn thành")
	status.Register(ExamReviewAlreadyExists, "This journey has already been reviewed", "Hành trình này đã có đánh giá")
	status.Register(ExamMissingExamID, "exam_id is required", "Thiếu mã đề (exam_id)")
	status.Register(ExamMissingIsVerify, "is_verify is required", "Thiếu trạng thái xác minh (is_verify)")
	status.Register(ExamPoolInvalidQuestions, "The corrected questions must keep the exam's question numbers and answer labels", "Bộ câu hỏi đã sửa phải giữ nguyên số câu và nhãn đáp án của đề")
	status.Register(ExamQuestionBankInsufficient, "The question bank does not hold enough suitable questions for this exam. Please try again later", "Ngân hàng câu hỏi chưa đủ câu hỏi phù hợp để tạo đề. Vui lòng thử lại sau")
	status.Register(ExamPoolBankReadonly, "This exam was built from the question bank; correct the questions in the bank instead", "Đề này lấy từ ngân hàng câu hỏi, hãy sửa câu hỏi trong ngân hàng")
}
