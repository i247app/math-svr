package err_svc

import "errors"

var (
	ErrTermNotFound           = errors.New("term not found")
	ErrTermMissingID          = errors.New("term ID is required")
	ErrTermMissingName        = errors.New("term name is required")
	ErrTermMissingDescription = errors.New("term description is required")
	ErrTermAlreadyExists      = errors.New("term already exists")
)
