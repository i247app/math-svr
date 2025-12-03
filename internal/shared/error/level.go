package err_svc

import "errors"

var (
	ErrLevelNotFound            = errors.New("grade not found")
	ErrLevelMissingLabel        = errors.New("grade label is required")
	ErrLevelMissingDescripton   = errors.New("grade description is required")
	ErrLevelMissingDisplayOrder = errors.New("grade display order is required")
	ErrLevelAlreadyExists       = errors.New("level already exists")
)
