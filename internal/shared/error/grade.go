package err_svc

import "errors"

var (
	ErrGradeNotFound = errors.New("grade not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrNotFound      = errors.New("resource not found")
)
