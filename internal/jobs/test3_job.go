package jobs

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

type TestJob3 struct {
}

func NewTestJob3() *TestJob3 { return &TestJob3{} }

const testJob3Name = "test.test_job_3"

func (j *TestJob3) Name() string { return testJob3Name }

func (j *TestJob3) Schedule() job.Schedule {
	return job.WeeklyAt(time.Wednesday, 22, 17, HoChiMinhTimezone)
}

func (j *TestJob3) Timeout() time.Duration { return 10 * time.Second }

func (j *TestJob3) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Infof("test_job_3.running %s", time.Now().Format(time.RFC3339))
	return nil
}
