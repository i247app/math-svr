package command

import "errors"

// Module-scoped sentinel errors for the school command package.
var (
	ErrSchoolNotFound             = errors.New("school not found")
	ErrSchoolNotFoundAfterInsert  = errors.New("school not found after insert")
	ErrSchoolNotFoundAfterUpdate  = errors.New("school not found after update")
)
