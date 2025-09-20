package ports

import (
	"context"
	"flux/internal/domain"
)

type JobService interface {
	CreateJob(ctx context.Context, job *domain.Job) error
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, queue string) ([]*domain.Job, error)
	ProcessJob(ctx context.Context, jobID string) error
}

type SchedulerService interface {
	SchedulePendingJobs(ctx context.Context) error
	ReclaimTimedOutJobs(ctx context.Context) error
	ProcessCronJobs(ctx context.Context) error
}

type WorkerService interface {
	ProcessJobs(ctx context.Context, queues []string) error
	Stop(ctx context.Context) error
}