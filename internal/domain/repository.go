package domain

import "context"

type JobRepository interface {
	CreateJob(ctx context.Context, job *Job) error
	GetJob(ctx context.Context, id string) (*Job, error)
	UpdateJob(ctx context.Context, job *Job) error
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, queue string, status JobStatus, limit int) ([]*Job, error)
	GetJobsReadyToRun(ctx context.Context, limit int) ([]*Job, error)
	GetScheduledJobs(ctx context.Context, limit int) ([]*Job, error)
	GetTimedOutJobs(ctx context.Context, limit int) ([]*Job, error)
	UpdateJobStatus(ctx context.Context, jobID string, status JobStatus) error
}

type JobRunRepository interface {
	CreateJobRun(ctx context.Context, run *JobRun) error
	GetJobRun(ctx context.Context, id string) (*JobRun, error)
	UpdateJobRun(ctx context.Context, run *JobRun) error
	ListJobRuns(ctx context.Context, jobID string) ([]*JobRun, error)
}

type WorkflowRepository interface {
	CreateWorkflow(ctx context.Context, workflow *Workflow) error
	GetWorkflow(ctx context.Context, id string) (*Workflow, error)
	UpdateWorkflow(ctx context.Context, workflow *Workflow) error
	DeleteWorkflow(ctx context.Context, id string) error
}

type WorkflowRunRepository interface {
	CreateWorkflowRun(ctx context.Context, run *WorkflowRun) error
	GetWorkflowRun(ctx context.Context, id string) (*WorkflowRun, error)
	UpdateWorkflowRun(ctx context.Context, run *WorkflowRun) error
}

