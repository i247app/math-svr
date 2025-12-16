package err_svc

import "errors"

var (
	ErrContactMissingName          = errors.New("name is required")
	ErrContactMissingMessage       = errors.New("message is required")
	ErrContactEmailOrPhoneRequired = errors.New("either email or phone is required")
	ErrContactInvalidPhone         = errors.New("invalid phone format")
	ErrContactInvalidEmail         = errors.New("invalid email format")
	ErrContactTooLongMessage       = errors.New("message too long")
	ErrContactTooLongName          = errors.New("name too long")
	ErrContactTooLongEmail         = errors.New("mail too long")
	ErrContactNotFound             = errors.New("contact not found")
	ErrContactAlreadyRead          = errors.New("contact already check read")
)
