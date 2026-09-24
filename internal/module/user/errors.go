package user

import "errors"

var (
	ErrAvatarFileRequired                        = errors.New("avatar file is required")
	ErrAvatarReferenceMustBeNonEmptyWhenProvided = errors.New("avatar reference must be non-empty when provided")
	ErrAvatarReferenceResolvedToEmptyKey         = errors.New("avatar reference resolved to empty key")
	ErrEmailRequired                             = errors.New("email is required")
	ErrNameRequired                              = errors.New("name is required")
	ErrPhoneRequired                             = errors.New("phone is required")
	ErrPhoneOrEmailRequired                      = errors.New("phone or email is required")
	ErrSessionNotFound                           = errors.New("session not found")
	ErrGuestSessionAlreadySecure                 = errors.New("this session already belongs to a registered account")
	ErrRoleInvalid                               = errors.New("role is invalid")
	ErrProvideEitherAvatarFileOrAvatarReference  = errors.New("provide either avatar file or avatar reference")
	ErrStorageAdapterNotConfigured               = errors.New("storage adapter is not configured")
	ErrUploadReturnedEmptyKey                    = errors.New("upload returned an empty key")
	ErrUserIDMustBeValidUUID                     = errors.New("user id must be a valid uuid")
	ErrUserNotFound                              = errors.New("user not found")
	ErrUserIDFormFieldRequired                   = errors.New("uid form field is required")
	ErrUserIDRequired                            = errors.New("uid is required")
	ErrUserNameMustBeNonEmptyWhenProvided        = errors.New("user_name must be non-empty when provided")
)
