package user

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrAvatarFileRequired                        = errors.New("avatar file is required")
	ErrAvatarReferenceMustBeNonEmptyWhenProvided = errors.New("avatar reference must be non-empty when provided")
	ErrAvatarReferenceResolvedToEmptyKey         = errors.New("avatar reference resolved to empty key")
	ErrEmailRequired                             = errors.New("email is required")
	ErrNameRequired                              = errors.New("name is required")
	ErrPhoneRequired                             = errors.New("phone is required")
	ErrRoleInvalid                               = errors.New("role is invalid")
	ErrProvideEitherAvatarFileOrAvatarReference  = errors.New("provide either avatar file or avatar reference")
	ErrStorageAdapterNotConfigured               = errors.New("storage adapter is not configured")
	ErrUploadReturnedEmptyKey                    = errors.New("upload returned an empty key")
	ErrUserIDMustBeValidUUID                     = errors.New("user id must be a valid uuid")
	ErrUserNotFound                              = errors.New("user not found")
	ErrUserIDFormFieldRequired                   = errors.New("user_id form field is required")
	ErrUserIDRequired                            = errors.New("user_id is required")
	ErrUserNameMustBeNonEmptyWhenProvided        = errors.New("user_name must be non-empty when provided")
)
