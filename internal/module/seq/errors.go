package seq

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrSeqNameRequired = errors.New("seq name is required")
)
