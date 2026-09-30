package command

import "errors"

// Module-scoped sentinel errors for the role command package.
var (
	ErrRoleNotFound            = errors.New("role not found")
	ErrRoleNotFoundAfterInsert = errors.New("role not found after insert")
	ErrRoleNotFoundAfterUpdate = errors.New("role not found after update")
	ErrRoleCodeAlreadyExists   = errors.New("role_code already exists")
)
