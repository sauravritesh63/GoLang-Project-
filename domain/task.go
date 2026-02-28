// Package domain contains the core business entities and interfaces
// for the distributed task scheduler.
package domain

import (
	"errors"
	"time"
)

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusRetrying  TaskStatus = "retrying"
)

// Priority controls the order in which tasks are dequeued.
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 5
	PriorityHigh   Priority = 10
)

// Task is the central domain entity representing a unit of work.
type Task struct {
	ID          string
	Name        string
	Payload     []byte
	Status      TaskStatus
	Priority    Priority
	MaxRetries  int
	RetryCount  int
	ScheduledAt time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Error       string
}

// Validate checks that a Task has the minimum required fields.
func (t *Task) Validate() error {
	if t.ID == "" {
		return errors.New("task ID must not be empty")
	}
	if t.Name == "" {
		return errors.New("task Name must not be empty")
	}
	if t.Priority < PriorityLow || t.Priority > PriorityHigh {
		return errors.New("task Priority must be between 1 and 10")
	}
	if t.MaxRetries < 0 {
		return errors.New("task MaxRetries must not be negative")
	}
	return nil
}

// CanRetry reports whether the task should be retried after a failure.
func (t *Task) CanRetry() bool {
	return t.RetryCount < t.MaxRetries
}

// IsTerminal reports whether the task has reached a final state.
func (t *Task) IsTerminal() bool {
	return t.Status == TaskStatusSucceeded || t.Status == TaskStatusFailed
}
