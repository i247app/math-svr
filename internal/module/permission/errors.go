package permission

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrRoleIDRequired              = errors.New("role_id is required")
	ErrRoleNotFound                = errors.New("role not found")
	ErrRoleCodeRequired            = errors.New("role_code is required")
	ErrInvalidRoleCode             = errors.New("invalid role_code")
	ErrRoleNameRequired            = errors.New("role_name is required")
	ErrRoleNameTooLong             = errors.New("role_name too long")
	ErrDescriptionTooLong          = errors.New("description too long")
	ErrNoteTooLong                 = errors.New("note too long")
	ErrInvalidRoleStatus           = errors.New("invalid role_status")
	ErrRoleImageConflict           = errors.New("send either role_image or remove_role_image, not both")
	ErrStorageAdapterNotConfigured = errors.New("storage adapter is not configured")
	ErrUploadReturnedEmptyKey      = errors.New("upload returned an empty key")
)
