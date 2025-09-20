package ports

import (
	"context"
	"flux/internal/domain"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *domain.Job) error
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	UpdateJob(ctx context.Context, job *domain.Job) error
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, queue string, status domain.JobStatus, limit int) ([]*domain.Job, error)
	GetJobsReadyToRun(ctx context.Context, limit int) ([]*domain.Job, error)
}

type JobRunRepository interface {
	CreateJobRun(ctx context.Context, run *domain.JobRun) error
	GetJobRun(ctx context.Context, id string) (*domain.JobRun, error)
	UpdateJobRun(ctx context.Context, run *domain.JobRun) error
	ListJobRuns(ctx context.Context, jobID string) ([]*domain.JobRun, error)
}

type WorkflowRepository interface {
	CreateWorkflow(ctx context.Context, workflow *domain.Workflow) error
	GetWorkflow(ctx context.Context, id string) (*domain.Workflow, error)
	UpdateWorkflow(ctx context.Context, workflow *domain.Workflow) error
	DeleteWorkflow(ctx context.Context, id string) error
}

type WorkflowRunRepository interface {
	CreateWorkflowRun(ctx context.Context, run *domain.WorkflowRun) error
	GetWorkflowRun(ctx context.Context, id string) (*domain.WorkflowRun, error)
	UpdateWorkflowRun(ctx context.Context, run *domain.WorkflowRun) error
}