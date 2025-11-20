package domain

import "context"

type JobManager struct{}

func NewJobManager(ctx context.Context) *JobManager {
	return &JobManager{}
}
