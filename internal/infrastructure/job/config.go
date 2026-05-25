package job

import "time"

// Config tunes the Runtime. A zero Config picks sensible defaults via
// withDefaults — every field is optional from the caller's perspective.
//
// All fields are read once in NewRuntime; later mutation is ignored.
type Config struct {
	// DefaultTimezone is the location handed to job constructors when
	// they want a project-default tz. The Runtime itself does not use
	// it (Schedule values carry their own *time.Location); it lives
	// here so internal/jobs/jobs.go has one source of truth.
	DefaultTimezone *time.Location

	// TaskWorkers is the size of the task worker pool. Default: 4.
	TaskWorkers int

	// TaskQueueSize is the bounded chan capacity. Once full, Enqueue
	// returns ErrQueueFull instead of blocking. Default: 1024.
	TaskQueueSize int

	// DrainTimeout is the wall-clock budget Stop waits for in-flight
	// executions to finish AFTER signalling cancellation. It does NOT
	// kill goroutines (Go has no safe primitive for that) — it logs a
	// warning if the drain breaches the budget. Default: 30s.
	DrainTimeout time.Duration

	// DefaultTaskTimeout caps a single task attempt when the caller
	// doesn't set TaskOptions.Timeout. Default: 60s.
	DefaultTaskTimeout time.Duration

	// DefaultRetryPolicy is applied when TaskOptions.Retry is the zero
	// value. Default: 3 attempts, exponential, base=5s, cap=5m.
	DefaultRetryPolicy RetryPolicy
}

func (c Config) withDefaults() Config {
	if c.TaskWorkers <= 0 {
		c.TaskWorkers = 4
	}
	if c.TaskQueueSize <= 0 {
		c.TaskQueueSize = 1024
	}
	if c.DrainTimeout <= 0 {
		c.DrainTimeout = 30 * time.Second
	}
	if c.DefaultTaskTimeout <= 0 {
		c.DefaultTaskTimeout = 60 * time.Second
	}
	if c.DefaultRetryPolicy == (RetryPolicy{}) {
		c.DefaultRetryPolicy = RetryPolicy{
			Attempts: 3,
			Backoff:  BackoffExponential,
			Base:     5 * time.Second,
			MaxDelay: 5 * time.Minute,
		}
	}
	return c
}
