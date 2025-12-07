package di

import "context"

type JobScheduler interface {
	Name() string
	TickInterval() int
	Run(ctx context.Context) error
}
