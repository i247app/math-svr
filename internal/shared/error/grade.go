package err_svc

import "errors"

var (
	ErrGradeNotFound            = errors.New("grade not found")
	ErrGradeMissingLabel        = errors.New("grade label is required")
	ErrGradeMissingDescripton   = errors.New("grade description is required")
	ErrGradeMissingDisplayOrder = errors.New("grade display order is required")
	ErrGradeAlreadyExists       = errors.New("grade already exists")
)
