package jobs

import (
	"context"
	"log"
	"time"
)

type TestJob struct {
	name string
	id   int
}

func NewTestJob() *TestJob {
	return &TestJob{
		name: "test-job",
		id:   30,
	}
}

func (j *TestJob) Name() string {
	return j.name
}

func (j *TestJob) TickInterval() int {
	return 30
}

func (j *TestJob) Run(ctx context.Context) error {
	now := time.Now()
	log.Printf("[%s] [id=%d] %s", j.name, j.id, now)
	return nil
}
