package command

import "errors"

// Module-scoped sentinel errors for the banner command package.
var (
	ErrBannerNotFound            = errors.New("banner not found")
	ErrBannerNotFoundAfterInsert = errors.New("banner not found after insert")
	ErrBannerNotFoundAfterUpdate = errors.New("banner not found after update")
)
