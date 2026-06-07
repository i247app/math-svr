package jobs

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// NoopJob is a hourly heartbeat that proves the scheduler is alive.
// Useful in production for "the job runtime is running" detection
// from log greps without needing metrics infra yet. It is intentionally
// always-registered — turn it off by pausing it via the control plane,
// not by deleting it.
type NoopJob struct{}

func NewNoopJob() *NoopJob { return &NoopJob{} }

const noopName = "system.noop"

func (j *NoopJob) Name() string { return noopName }
func (j *NoopJob) Schedule() job.Schedule {
	return job.DailyAt(22, 52, LosAngelesTimezone)
}

func (j *NoopJob) Timeout() time.Duration { return 10 * time.Second }

func (j *NoopJob) Run(ctx context.Context) error {
	logger.From(ctx).Info("system.noop.heartbeat")
	return nil
}
