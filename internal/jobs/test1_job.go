package jobs

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

type TestJob1 struct {
}

func NewTestJob1() *TestJob1 { return &TestJob1{} }

const testJob1Name = "test.test_job_1"

func (j *TestJob1) Name() string { return testJob1Name }

func (j *TestJob1) Schedule() job.Schedule {
	return job.EveryDuration(10 * time.Minute)
}

func (j *TestJob1) Timeout() time.Duration { return 10 * time.Second }

func (j *TestJob1) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Infof("test_job_1.running %s", time.Now().Format(time.RFC3339))
	return nil
}
