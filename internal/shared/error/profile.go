package err_svc

import "errors"

var (
	ErrProfileMissingGrade  = errors.New("grade is required for profile")
	ErrProfileMissingTerm   = errors.New("term is required for profile")
	ErrProfileAlreadyExists = errors.New("profile already exists")
)
