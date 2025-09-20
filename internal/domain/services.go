package domain

import "context"

type SchedulerService interface {
	SchedulePendingJobs(ctx context.Context) error
	ReclaimTimedOutJobs(ctx context.Context) error
	ProcessCronJobs(ctx context.Context) error
}

type WorkerService interface {
	ProcessJobs(ctx context.Context, queues []string) error
	Stop(ctx context.Context) error
}