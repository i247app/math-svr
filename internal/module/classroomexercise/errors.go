package classroomexercise

import "errors"

// Module-scoped sentinel errors for the classroom-exercise presentation
// package. See internal/module/classroom/errors.go for the surrounding
// convention.
var (
	// Validation — bounds & required fields.
	ErrTitleRequired       = errors.New("title is required")
	ErrTitleTooLong        = errors.New("title too long")
	ErrChapterNameRequired = errors.New("chapter_name is required")
	ErrChapterNameTooLong  = errors.New("chapter_name too long")
	ErrLessonNameRequired  = errors.New("lesson_name is required")
	ErrLessonNameTooLong   = errors.New("lesson_name too long")
	ErrDescriptionTooLong  = errors.New("description too long")
	ErrNoteTooLong         = errors.New("note too long")
	ErrSearchTooLong       = errors.New("search too long")
	ErrNumQuestionsOutOfRange = errors.New("num_questions out of range")
	ErrEndDateBeforeStart  = errors.New("end_date must be after start_date")
	ErrInvalidStatus       = errors.New("invalid status")
	ErrInvalidExerciseStatus = errors.New("invalid exercise_status")
	ErrInvalidVisibility   = errors.New("invalid visibility")
	ErrInvalidSortBy       = errors.New("invalid sort_by")
	ErrInvalidSortOrder    = errors.New("invalid sort_order")
	ErrClassroomExerciseIDRequired = errors.New("classroom_exercise_id is required")
	ErrClassroomIDRequired = errors.New("classroom_id is required")
	ErrProfileIDRequired   = errors.New("profile_id is required")
	ErrProfileIDRequiredMultiProfile = errors.New("profile_id is required when the user owns multiple profiles")
	ErrNilRequest          = errors.New("nil request")

	// Cross-aggregate lookups.
	ErrClassroomNotFound        = errors.New("classroom not found")
	ErrClassroomExerciseNotFound = errors.New("classroom exercise not found")
	ErrProfileNotFound          = errors.New("profile not found")

	// Permission / business.
	ErrNotClassroomMember        = errors.New("not a member of this classroom")
	ErrMembershipNotActive       = errors.New("membership is not active")
	ErrManagerRoleRequired       = errors.New("manager role required")
	ErrPrivateExerciseOwnerOnly  = errors.New("private exercise — only the creator can access")
	ErrProgramNotAssociated      = errors.New("program is not associated with this classroom")

	// Session / acting-profile resolution.
	ErrUIDNotFoundFromSession  = errors.New("uid not found from session")
	ErrProfileNotOwnedByUser   = errors.New("profile does not belong to the authenticated user")

	// Adapter wiring.
	ErrBotAdapterNotConfigured = errors.New("classroom exercise: bot adapter is not configured")
)
