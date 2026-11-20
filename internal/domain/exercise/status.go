package exercise

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (12600–12799). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Classroom exercise — teacher-issued, AI-generated assignment
	// scoped to a classroom. Curriculum context (grade, program) is
	// derived from the classroom; chapter_name + lesson_name are the
	// teacher-supplied topic anchors fed to the llm.
	ClassroomExerciseNotFound              status.StatusCode = 12601
	ClassroomExerciseMissingID             status.StatusCode = 12602
	ClassroomExerciseMissingClassroomID    status.StatusCode = 12603
	ClassroomExerciseMissingTitle          status.StatusCode = 12604
	ClassroomExerciseTitleTooLong          status.StatusCode = 12605
	ClassroomExerciseMissingChapterName    status.StatusCode = 12606
	ClassroomExerciseChapterNameTooLong    status.StatusCode = 12607
	ClassroomExerciseMissingLessonName     status.StatusCode = 12608
	ClassroomExerciseLessonNameTooLong     status.StatusCode = 12609
	ClassroomExerciseInvalidNumQuestions   status.StatusCode = 12610
	ClassroomExerciseInvalidDateRange      status.StatusCode = 12611
	ClassroomExerciseNoteTooLong           status.StatusCode = 12612
	ClassroomExerciseInvalidProgram        status.StatusCode = 12613
	ClassroomExerciseProgramNotInClassroom status.StatusCode = 12614
	ClassroomExerciseAlreadyDeleted        status.StatusCode = 12615
	ClassroomExercisePermissionDenied      status.StatusCode = 12616
	ClassroomExerciseGenerationFailed      status.StatusCode = 12617
	ClassroomExerciseInvalidVisibility     status.StatusCode = 12618
	ClassroomExercisePrivateDenied         status.StatusCode = 12619
	ClassroomExerciseInvalidPurpose        status.StatusCode = 12620
	// Classroom exercise submissions — 12700-12799
	ClassroomExerciseSubmissionNotFound            status.StatusCode = 12700
	ClassroomExerciseSubmissionMissingID           status.StatusCode = 12701
	ClassroomExerciseSubmissionMissingExerciseID   status.StatusCode = 12702
	ClassroomExerciseSubmissionMissingAnswers      status.StatusCode = 12703
	ClassroomExerciseSubmissionInvalidAnswers      status.StatusCode = 12704
	ClassroomExerciseSubmissionAlreadyExists       status.StatusCode = 12705
	ClassroomExerciseSubmissionWindowNotOpen       status.StatusCode = 12706
	ClassroomExerciseSubmissionWindowClosed        status.StatusCode = 12707
	ClassroomExerciseSubmissionExerciseUnavailable status.StatusCode = 12708
	ClassroomExerciseSubmissionPermissionDenied    status.StatusCode = 12709
	ClassroomExerciseSubmissionGradingFailed       status.StatusCode = 12710
	ClassroomExerciseSubmissionNoteTooLong         status.StatusCode = 12711
	ClassroomExerciseSubmissionAlreadyDeleted      status.StatusCode = 12712
)

func init() {
	status.Register(ClassroomExerciseNotFound, "Classroom exercise not found", "Không tìm thấy bài tập lớp học")
	status.Register(ClassroomExerciseMissingID, "Classroom exercise id is required", "Mã bài tập là bắt buộc")
	status.Register(ClassroomExerciseMissingClassroomID, "Classroom id is required", "Mã lớp học là bắt buộc")
	status.Register(ClassroomExerciseMissingTitle, "Exercise title is required", "Tên bài tập là bắt buộc")
	status.Register(ClassroomExerciseTitleTooLong, "Exercise title is too long", "Tên bài tập quá dài")
	status.Register(ClassroomExerciseMissingChapterName, "Chapter name is required", "Tên chương là bắt buộc")
	status.Register(ClassroomExerciseChapterNameTooLong, "Chapter name is too long", "Tên chương quá dài")
	status.Register(ClassroomExerciseMissingLessonName, "Lesson name is required", "Tên bài học là bắt buộc")
	status.Register(ClassroomExerciseLessonNameTooLong, "Lesson name is too long", "Tên bài học quá dài")
	status.Register(ClassroomExerciseInvalidNumQuestions, "Number of questions is invalid", "Số lượng câu hỏi không hợp lệ")
	status.Register(ClassroomExerciseInvalidDateRange, "Exercise end date must be after start date", "Ngày kết thúc phải sau ngày bắt đầu")
	status.Register(ClassroomExerciseNoteTooLong, "Exercise note is too long", "Ghi chú quá dài")
	status.Register(ClassroomExerciseInvalidProgram, "Program is invalid", "Bộ sách không hợp lệ")
	status.Register(ClassroomExerciseProgramNotInClassroom, "Program is not associated with this classroom", "Bộ sách không thuộc lớp học này")
	status.Register(ClassroomExerciseAlreadyDeleted, "Classroom exercise is already deleted", "Bài tập đã bị xoá")
	status.Register(ClassroomExercisePermissionDenied, "You do not have permission to manage this exercise", "Bạn không có quyền thao tác trên bài tập này")
	status.Register(ClassroomExerciseGenerationFailed, "Failed to generate classroom exercise", "Tạo bài tập tự động thất bại")
	status.Register(ClassroomExerciseInvalidVisibility, "Visibility must be PUBLIC or PRIVATE", "Chế độ hiển thị phải là PUBLIC hoặc PRIVATE")
	status.Register(ClassroomExercisePrivateDenied, "Only the creator can access this private exercise", "Chỉ người tạo mới có quyền truy cập bài tập riêng tư này")
	status.Register(ClassroomExerciseInvalidPurpose, "Purpose must be HOMEWORK or EXAM", "Mục đích phải là HOMEWORK hoặc EXAM")
	status.Register(ClassroomExerciseSubmissionNotFound, "Submission not found", "Không tìm thấy bài nộp")
	status.Register(ClassroomExerciseSubmissionMissingID, "classroom_exercise_submission_id is required", "classroom_exercise_submission_id là bắt buộc")
	status.Register(ClassroomExerciseSubmissionMissingExerciseID, "classroom_exercise_id is required", "classroom_exercise_id là bắt buộc")
	status.Register(ClassroomExerciseSubmissionMissingAnswers, "answers is required", "answers là bắt buộc")
	status.Register(ClassroomExerciseSubmissionInvalidAnswers, "Invalid answers payload", "Dữ liệu câu trả lời không hợp lệ")
	status.Register(ClassroomExerciseSubmissionAlreadyExists, "You have already submitted this exercise", "Bạn đã nộp bài tập này rồi")
	status.Register(ClassroomExerciseSubmissionWindowNotOpen, "Submission window has not opened yet", "Bài tập chưa mở để làm")
	status.Register(ClassroomExerciseSubmissionWindowClosed, "Submission window has closed", "Bài tập đã hết hạn nộp")
	status.Register(ClassroomExerciseSubmissionExerciseUnavailable, "Exercise is not available for submission", "Bài tập hiện không khả dụng để nộp")
	status.Register(ClassroomExerciseSubmissionPermissionDenied, "You do not have permission to access this submission", "Bạn không có quyền truy cập bài nộp này")
	status.Register(ClassroomExerciseSubmissionGradingFailed, "Failed to grade the submission", "Chấm bài tự động thất bại")
	status.Register(ClassroomExerciseSubmissionNoteTooLong, "Note too long", "Ghi chú quá dài")
	status.Register(ClassroomExerciseSubmissionAlreadyDeleted, "Submission is already deleted", "Bài nộp đã bị xoá")
}
