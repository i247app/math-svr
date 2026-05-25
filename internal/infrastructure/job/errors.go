package job

import "errors"

// Sentinel errors returned by the Runtime. The module layer translates
// these into MathError values with the matching JOB_* status code; the
// runtime itself stays free of presentation concerns.
var (
	ErrJobNotFound         = errors.New("job: not found")
	ErrJobAlreadyPaused    = errors.New("job: already paused")
	ErrJobNotPaused        = errors.New("job: not paused")
	ErrJobInFlight         = errors.New("job: an execution is already in flight")
	ErrRuntimeUnavailable  = errors.New("job: runtime not started or already stopping")
	ErrTaskHandlerNotFound = errors.New("job: task handler not registered")
	ErrQueueFull           = errors.New("job: task queue is full")
)
