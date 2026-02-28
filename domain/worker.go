package domain

import (
	"errors"
	"time"
)

// WorkerStatus represents the availability of a worker node.
type WorkerStatus string

const (
	WorkerStatusIdle    WorkerStatus = "idle"
	WorkerStatusBusy    WorkerStatus = "busy"
	WorkerStatusDrained WorkerStatus = "drained"
	WorkerStatusOffline WorkerStatus = "offline"
)

// Worker represents a node that executes tasks.
type Worker struct {
	ID          string
	Address     string
	Status      WorkerStatus
	Concurrency int
	ActiveTasks int
	LastHeartAt time.Time
	RegisteredAt time.Time
}

// Validate checks that a Worker has the minimum required fields.
func (w *Worker) Validate() error {
	if w.ID == "" {
		return errors.New("worker ID must not be empty")
	}
	if w.Address == "" {
		return errors.New("worker Address must not be empty")
	}
	if w.Concurrency <= 0 {
		return errors.New("worker Concurrency must be greater than zero")
	}
	return nil
}

// HasCapacity reports whether the worker can accept another task.
func (w *Worker) HasCapacity() bool {
	return (w.Status == WorkerStatusIdle) ||
		(w.Status == WorkerStatusBusy && w.ActiveTasks < w.Concurrency)
}

// IsAlive reports whether a recent heartbeat has been received within the given
// timeout window.
func (w *Worker) IsAlive(timeout time.Duration) bool {
	return time.Since(w.LastHeartAt) <= timeout
}
