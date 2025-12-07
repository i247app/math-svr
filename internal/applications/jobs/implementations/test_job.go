package implementations

import (
	"context"
	"log"
	"time"

	di "math-ai.com/math-ai/internal/core/di/jobs"
)

// TestJob is a simple test job that demonstrates the job pattern
// It doesn't require any dependencies
type TestJob struct {
	name         string
	tickInterval int
}

// NewTestJob creates a new test job
func NewTestJob() di.JobScheduler {
	return &TestJob{
		name:         "test-job",
		tickInterval: 30, // Run every 30 minutes
	}
}

// Name returns the job name
func (j *TestJob) Name() string {
	return j.name
}

// TickInterval returns the job tick interval in minutes
func (j *TestJob) TickInterval() int {
	return j.tickInterval
}

// Run executes the test job
func (j *TestJob) Run(ctx context.Context) error {
	now := time.Now()
	log.Printf("[JOB] [%s] Executing at: %s", j.name, now.Format(time.RFC3339))

	// Example job logic
	// This is where you would put your actual job work

	return nil
}
