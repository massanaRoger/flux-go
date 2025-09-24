package domain

import "context"

type JobRepository interface {
	// CreateJob persists a new job to the repository
	CreateJob(ctx context.Context, job *Job) error
	// GetJob retrieves a job by its unique identifier
	GetJob(ctx context.Context, id string) (*Job, error)
	// UpdateJob updates an existing job's data in the repository
	UpdateJob(ctx context.Context, job *Job) error
	// DeleteJob removes a job from the repository
	DeleteJob(ctx context.Context, id string) error
	// ListJobs retrieves jobs filtered by queue and status with a limit
	ListJobs(ctx context.Context, queue string, status JobStatus, limit int) ([]*Job, error)
	// GetJobsReadyToRun retrieves jobs that are ready for execution
	GetJobsReadyToRun(ctx context.Context, limit int) ([]*Job, error)
	// GetScheduledJobs retrieves jobs that are scheduled for future execution
	GetScheduledJobs(ctx context.Context, limit int) ([]*Job, error)
	// GetTimedOutJobs retrieves jobs that have exceeded their execution timeout
	GetTimedOutJobs(ctx context.Context, limit int) ([]*Job, error)
	// UpdateJobStatus updates only the status of a specific job
	UpdateJobStatus(ctx context.Context, jobID string, status JobStatus) error
}

type JobRunRepository interface {
	// CreateJobRun persists a new job execution record to the repository
	CreateJobRun(ctx context.Context, run *JobRun) error
	// GetJobRun retrieves a job run by its unique identifier
	GetJobRun(ctx context.Context, id string) (*JobRun, error)
	// UpdateJobRun updates an existing job run's data in the repository
	UpdateJobRun(ctx context.Context, run *JobRun) error
	// ListJobRuns retrieves all job runs for a specific job
	ListJobRuns(ctx context.Context, jobID string) ([]*JobRun, error)
}

type WorkflowRepository interface {
	// CreateWorkflow persists a new workflow definition to the repository
	CreateWorkflow(ctx context.Context, workflow *Workflow) error
	// GetWorkflow retrieves a workflow by its unique identifier
	GetWorkflow(ctx context.Context, id string) (*Workflow, error)
	// UpdateWorkflow updates an existing workflow's data in the repository
	UpdateWorkflow(ctx context.Context, workflow *Workflow) error
	// DeleteWorkflow removes a workflow from the repository
	DeleteWorkflow(ctx context.Context, id string) error
}

type WorkflowRunRepository interface {
	// CreateWorkflowRun persists a new workflow execution record to the repository
	CreateWorkflowRun(ctx context.Context, run *WorkflowRun) error
	// GetWorkflowRun retrieves a workflow run by its unique identifier
	GetWorkflowRun(ctx context.Context, id string) (*WorkflowRun, error)
	// UpdateWorkflowRun updates an existing workflow run's data in the repository
	UpdateWorkflowRun(ctx context.Context, run *WorkflowRun) error
}

