package err_svc

import "errors"

var (
	ErrSemesterNotFound           = errors.New("semester not found")
	ErrSemesterMissingName        = errors.New("semester name is required")
	ErrSemesterMissingDescription = errors.New("semester description is required")
	ErrSemesterAlreadyExists      = errors.New("semester already exists")
)
