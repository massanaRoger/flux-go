package domain

import (
	"encoding/json"
	"time"
)

type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusSucceeded WorkflowStatus = "succeeded"
	WorkflowStatusFailed    WorkflowStatus = "failed"
)

type Workflow struct {
	ID         string
	Name       string
	Definition json.RawMessage
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type WorkflowRun struct {
	ID         string
	WorkflowID string
	Status     WorkflowStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}