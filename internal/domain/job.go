package domain

import (
	"encoding/json"
	"time"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusScheduled JobStatus = "scheduled"
	JobStatusRunning   JobStatus = "running"
	JobStatusSucceeded JobStatus = "succeeded"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
)

type Job struct {
	ID                   string
	Type                 string
	Queue                string
	Payload              json.RawMessage
	Priority             int
	Status               JobStatus
	ScheduleAt           time.Time
	TimeoutSec           int
	Attempts             int
	MaxRetries           int
	BackoffPolicy        string
	VisibilityTimeoutSec int
	IdempotencyKey       *string
	LastError            *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type JobRun struct {
	ID         string
	JobID      string
	StartedAt  time.Time
	FinishedAt *time.Time
	Status     JobStatus
	Error      *string
	Metrics    json.RawMessage
}

func NewJob(jobType, queue string, payload json.RawMessage) *Job {
	return &Job{
		Type:                 jobType,
		Queue:                queue,
		Payload:              payload,
		Priority:             0,
		Status:               JobStatusPending,
		ScheduleAt:           time.Now(),
		TimeoutSec:           300,
		Attempts:             0,
		MaxRetries:           3,
		BackoffPolicy:        "exp",
		VisibilityTimeoutSec: 60,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}