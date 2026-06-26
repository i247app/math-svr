package classroom

import "errors"

var (
	// Validation — bounds & required fields.
	ErrNameRequired                     = errors.New("name is required")
	ErrNameTooLong                      = errors.New("name too long")
	ErrNameBlank                        = errors.New("name cannot be blank")
	ErrDescriptionTooLong               = errors.New("description too long")
	ErrNoteTooLong                      = errors.New("note too long")
	ErrCoverKeyTooLong                  = errors.New("cover_key too long")
	ErrSearchTooLong                    = errors.New("search too long")
	ErrMaxMembersInvalid                = errors.New("max_members must be > 0")
	ErrClassroomCodeRequired            = errors.New("classroom_code is required")
	ErrClassroomCodeTooLong             = errors.New("classroom_code too long")
	ErrClassroomCodeExpiresMustBeFuture = errors.New("classroom_code_expires_dt must be in the future")
	ErrClassroomIDRequired              = errors.New("classroom_id is required")
	ErrProfileIDRequired                = errors.New("profile_id is required")
	ErrProfileIDRequiredNoSession       = errors.New("profile_id is required when no session is attached")
	ErrTargetProfileIDRequired          = errors.New("target_profile_id is required")
	ErrNewOwnerProfileIDRequired        = errors.New("new_owner_profile_id is required")
	ErrNewRoleRequired                  = errors.New("new_role is required")
	ErrNewRoleInvalid                   = errors.New("new_role must be CO_TEACHER or STUDENT")
	ErrAtLeastOneTargetRequired         = errors.New("at least one target is required")
	ErrDuplicateProgramIDInList         = errors.New("duplicate program_id in list")

	// Cross-aggregate refs missing.
	ErrSchoolNotFound          = errors.New("school not found")
	ErrProgramNotFound         = errors.New("program not found")
	ErrGradeNotFound           = errors.New("grade not found")
	ErrTargetProfileNotFound   = errors.New("target profile not found")
	ErrNewOwnerProfileNotFound = errors.New("new owner profile not found")
	ErrProfileNotFound         = errors.New("profile not found")
	ErrClassroomNotFound       = errors.New("classroom not found")
	ErrTargetMemberNotFound    = errors.New("target member not found")

	// Business / permission.
	ErrClassroomArchived             = errors.New("classroom is archived")
	ErrNotClassroomMember            = errors.New("not a member of this classroom")
	ErrMembershipNotActive           = errors.New("membership is not active")
	ErrManagerRoleRequired           = errors.New("manager role required")
	ErrOwnerRoleRequired             = errors.New("owner role required")
	ErrTeacherOwnershipOnly          = errors.New("only a teacher profile can own a classroom")
	ErrTeacherCoTeacherOnly          = errors.New("only teacher profiles can become co-teachers")
	ErrCoTeacherCanOnlyRemoveStudent = errors.New("co-teacher can only remove students")
	ErrCannotRemoveSelf              = errors.New("cannot remove yourself; use leave instead")
	ErrCannotTransferToSelf          = errors.New("cannot transfer ownership to yourself")

	// Session / acting-profile resolution.
	ErrUIDNotFoundFromSession = errors.New("uid not found from session")
	ErrNoProfileForUser       = errors.New("no profile found for the authenticated user")
	ErrProfileNotOwnedByUser  = errors.New("profile does not belong to the authenticated user")

	// Class-learning-progress validators.
	ErrProgressInvalidBucket    = errors.New("bucket must be DAY, WEEK, or MONTH")
	ErrProgressInvalidDateRange = errors.New("from_dt must be before to_dt")
	ErrProgressRangeTooLong     = errors.New("date range exceeds the 2-year cap")
	ErrProgressInvalidTz        = errors.New("tz must be a numeric offset like +07:00")
	ErrProgressInvalidPurpose   = errors.New("purpose must be HOMEWORK, QUIZ, or EXAM")
)
