package job

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrJobNameRequired          = errors.New("job name is required")
	ErrJobRuntimeNotInitialised = errors.New("job runtime is not initialised")
	ErrTaskNameRequired         = errors.New("task name is required")
)
