// Package repository defines the persistence interfaces for the distributed
// task scheduler. All methods are context-aware and accept/return domain types.
// Concrete implementations live in sub-packages (e.g. postgres, mock).
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
)

// WorkflowRepository defines CRUD and query operations for Workflow entities.
type WorkflowRepository interface {
	// Create persists a new workflow. The caller is responsible for setting wf.ID.
	Create(ctx context.Context, wf *domain.Workflow) error
	// GetByID returns the workflow with the given ID, or ErrNotFound.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Workflow, error)
	// Update overwrites all mutable fields of an existing workflow.
	Update(ctx context.Context, wf *domain.Workflow) error
	// Delete removes the workflow record.
	Delete(ctx context.Context, id uuid.UUID) error
	// List returns all workflows ordered by creation time (newest first).
	List(ctx context.Context) ([]*domain.Workflow, error)
	// ListActive returns only workflows where is_active = true.
	ListActive(ctx context.Context) ([]*domain.Workflow, error)
}

// TaskRepository defines CRUD and query operations for Task entities.
type TaskRepository interface {
	// Create persists a new task. The caller is responsible for setting t.ID.
	Create(ctx context.Context, t *domain.Task) error
	// GetByID returns the task with the given ID, or ErrNotFound.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	// Update overwrites all mutable fields of an existing task.
	Update(ctx context.Context, t *domain.Task) error
	// Delete removes the task record.
	Delete(ctx context.Context, id uuid.UUID) error
	// ListByWorkflowID returns all tasks belonging to the given workflow,
	// ordered by creation time (oldest first).
	ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.Task, error)
}

// WorkflowRunRepository defines CRUD and query operations for WorkflowRun entities.
type WorkflowRunRepository interface {
	// Create persists a new workflow run. The caller is responsible for setting wr.ID.
	Create(ctx context.Context, wr *domain.WorkflowRun) error
	// GetByID returns the run with the given ID, or ErrNotFound.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowRun, error)
	// UpdateStatus atomically updates the status and optional finished timestamp.
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, finishedAt *time.Time) error
	// ListByWorkflowID returns all runs for the given workflow, newest first.
	ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.WorkflowRun, error)
	// ListByStatus returns all runs with the given status, newest first.
	ListByStatus(ctx context.Context, status domain.Status) ([]*domain.WorkflowRun, error)
}

// TaskRunRepository defines CRUD and query operations for TaskRun entities.
type TaskRunRepository interface {
	// Create persists a new task run. The caller is responsible for setting tr.ID.
	Create(ctx context.Context, tr *domain.TaskRun) error
	// GetByID returns the task run with the given ID, or ErrNotFound.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TaskRun, error)
	// UpdateStatus atomically updates the status and optional finished timestamp.
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, finishedAt *time.Time) error
	// ListByWorkflowRunID returns all task runs belonging to the given workflow run.
	ListByWorkflowRunID(ctx context.Context, workflowRunID uuid.UUID) ([]*domain.TaskRun, error)
	// ListByTaskID returns all runs for a specific task definition across all workflow runs.
	ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskRun, error)
	// ListByStatus returns all task runs with the given status.
	ListByStatus(ctx context.Context, status domain.Status) ([]*domain.TaskRun, error)
}

// WorkerRepository defines CRUD and query operations for Worker entities.
type WorkerRepository interface {
	// Create persists a new worker registration. The caller is responsible for setting w.ID.
	Create(ctx context.Context, w *domain.Worker) error
	// GetByID returns the worker with the given ID, or ErrNotFound.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Worker, error)
	// Update overwrites all mutable fields of an existing worker.
	Update(ctx context.Context, w *domain.Worker) error
	// Delete removes the worker record.
	Delete(ctx context.Context, id uuid.UUID) error
	// ListActive returns all workers with status = WorkerStatusActive.
	ListActive(ctx context.Context) ([]*domain.Worker, error)
	// UpdateHeartbeat sets last_heartbeat to at for the given worker.
	UpdateHeartbeat(ctx context.Context, id uuid.UUID, at time.Time) error
}

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errNotFound("record not found")

type errNotFound string

func (e errNotFound) Error() string { return string(e) }
