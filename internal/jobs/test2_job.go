package jobs

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

type TestJob2 struct {
}

func NewTestJob2() *TestJob2 { return &TestJob2{} }

const testJob2Name = "test.test_job_2"

func (j *TestJob2) Name() string { return testJob2Name }

func (j *TestJob2) Schedule() job.Schedule {
	return job.DailyAt(22, 00, HoChiMinhTimezone)
}

func (j *TestJob2) Timeout() time.Duration { return 10 * time.Second }

func (j *TestJob2) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Infof("test_job_2.running %s", time.Now().Format(time.RFC3339))
	return nil
}
