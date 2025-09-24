package domain

import "context"

type SchedulerService interface {
	// SchedulePendingJobs finds pending jobs and schedules them for execution
	SchedulePendingJobs(ctx context.Context) error
	// ReclaimTimedOutJobs identifies jobs that have exceeded their timeout and marks them for retry
	ReclaimTimedOutJobs(ctx context.Context) error
	// ProcessCronJobs evaluates cron expressions and schedules recurring jobs
	ProcessCronJobs(ctx context.Context) error
}

type JobHandler func(context.Context, *QueueMessage) error

type WorkerService interface {
	// Register job handlers for different job types
	RegisterHandler(jobType string, handler JobHandler) error
	// ProcessJobs continuously processes jobs from the specified queues
	ProcessJobs(queues []string) error
	// Stop gracefully shuts down the worker service
	Stop(ctx context.Context) error
}

