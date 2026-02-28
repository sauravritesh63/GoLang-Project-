package domain

import "context"

// TaskRepository defines the persistence operations for Tasks.
type TaskRepository interface {
	// Save creates or updates a task.
	Save(ctx context.Context, task *Task) error
	// FindByID returns the task with the given ID or ErrTaskNotFound.
	FindByID(ctx context.Context, id string) (*Task, error)
	// FindByStatus returns all tasks in the given status, ordered by priority
	// (highest first) and then by ScheduledAt (earliest first).
	FindByStatus(ctx context.Context, status TaskStatus) ([]*Task, error)
	// Delete removes the task record.
	Delete(ctx context.Context, id string) error
}

// WorkerRepository defines the persistence operations for Workers.
type WorkerRepository interface {
	// Save creates or updates a worker registration.
	Save(ctx context.Context, worker *Worker) error
	// FindByID returns the worker with the given ID or ErrWorkerNotFound.
	FindByID(ctx context.Context, id string) (*Worker, error)
	// FindAvailable returns all workers that currently have capacity.
	FindAvailable(ctx context.Context) ([]*Worker, error)
	// Delete removes the worker record.
	Delete(ctx context.Context, id string) error
}

// Queue defines the operations for the distributed task queue.
type Queue interface {
	// Enqueue pushes a task onto the queue.
	Enqueue(ctx context.Context, task *Task) error
	// Dequeue blocks until a task is available and returns it, or returns an
	// error if the context is cancelled.
	Dequeue(ctx context.Context) (*Task, error)
	// Len returns the current depth of the queue.
	Len(ctx context.Context) (int, error)
}

// Scheduler defines the high-level scheduling operations.
type Scheduler interface {
	// Submit accepts a new task and enqueues it for execution.
	Submit(ctx context.Context, task *Task) error
	// Cancel marks an in-flight or queued task as failed without executing it.
	Cancel(ctx context.Context, taskID string) error
	// Status returns the current state of a task.
	Status(ctx context.Context, taskID string) (TaskStatus, error)
}
