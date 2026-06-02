package profile

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrAvatarFileRequired                        = errors.New("avatar file is required")
	ErrAvatarReferenceMustBeNonEmptyWhenProvided = errors.New("avatar reference must be non-empty when provided")
	ErrAvatarReferenceResolvedToEmptyKey         = errors.New("avatar reference resolved to empty key")
	ErrGradeIDRequired                           = errors.New("grade_id is required")
	ErrLanguageMustBeVnOrEn                      = errors.New("language must be 'vn' or 'en'")
	ErrNameCannotBeBlank                         = errors.New("name cannot be blank")
	ErrNameRequired                              = errors.New("name is required")
	ErrProfileNotFound                           = errors.New("profile not found")
	ErrProfileIDFormFieldRequired                = errors.New("profile_id form field is required")
	ErrProfileIDRequired                         = errors.New("profile_id is required")
	ErrProfileStatusInvalid                      = errors.New("profile_status is invalid")
	ErrProgramIDRequired                         = errors.New("program_id is required")
	ErrProvideEitherAvatarFileOrAvatarReference  = errors.New("provide either avatar file or avatar reference")
	ErrRoleInvalid                               = errors.New("role is invalid")
	ErrSchoolIDRequired                          = errors.New("school_id is required")
	ErrSearchTooLong                             = errors.New("search too long")
	ErrSemesterIDRequired                        = errors.New("semester_id is required")
	ErrStorageAdapterNotConfigured               = errors.New("storage adapter is not configured")
	ErrUploadReturnedEmptyKey                    = errors.New("upload returned an empty key")
	ErrUserIDRequired                            = errors.New("user_id is required")
)
