package job

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrJobNameRequired          = errors.New("job name is required")
	ErrJobRuntimeNotInitialised = errors.New("job runtime is not initialised")
	ErrTaskNameRequired         = errors.New("task name is required")

	ErrScheduleRequired    = errors.New("schedule is required")
	ErrScheduleKindUnknown = errors.New(`schedule kind must be one of "every", "daily", "weekly"`)
	ErrWeekdayRequired     = errors.New("weekday is required for a weekly schedule (0=Sunday … 6=Saturday)")
)
