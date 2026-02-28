// Package domain defines the core data models for the distributed task scheduler (Mini Airflow).
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle state of a workflow run or task run.
type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

// WorkerStatus represents the availability state of a worker node.
type WorkerStatus string

const (
	WorkerStatusActive   WorkerStatus = "active"
	WorkerStatusInactive WorkerStatus = "inactive"
)

// Workflow is a named, schedulable collection of tasks.
type Workflow struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ScheduleCron string    `json:"schedule_cron"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// Task is a single unit of work that belongs to a Workflow.
type Task struct {
	ID                uuid.UUID `json:"id"`
	WorkflowID        uuid.UUID `json:"workflow_id"`
	Name              string    `json:"name"`
	Command           string    `json:"command"`
	RetryCount        int       `json:"retry_count"`
	RetryDelaySeconds int       `json:"retry_delay_seconds"`
	TimeoutSeconds    int       `json:"timeout_seconds"`
	CreatedAt         time.Time `json:"created_at"`
}

// TaskDependency records that a task must wait for another task to complete first.
type TaskDependency struct {
	ID              uuid.UUID `json:"id"`
	TaskID          uuid.UUID `json:"task_id"`
	DependsOnTaskID uuid.UUID `json:"depends_on_task_id"`
}

// WorkflowRun is a single execution instance of a Workflow.
type WorkflowRun struct {
	ID         uuid.UUID  `json:"id"`
	WorkflowID uuid.UUID  `json:"workflow_id"`
	Status     Status     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// TaskRun is a single execution attempt of a Task within a WorkflowRun.
type TaskRun struct {
	ID            uuid.UUID  `json:"id"`
	WorkflowRunID uuid.UUID  `json:"workflow_run_id"`
	TaskID        uuid.UUID  `json:"task_id"`
	Status        Status     `json:"status"`
	Attempt       int        `json:"attempt"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	Logs          string     `json:"logs"`
}

// Worker represents a node that picks up and executes tasks.
type Worker struct {
	ID            uuid.UUID    `json:"id"`
	Hostname      string       `json:"hostname"`
	LastHeartbeat time.Time    `json:"last_heartbeat"`
	Status        WorkerStatus `json:"status"`
}
