package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/sauravritesh63/GoLang-Project-/domain"
)

// Scheduler implements domain.Scheduler. It validates and enqueues tasks,
// tracks their status via the TaskRepository, and supports cancellation.
type Scheduler struct {
	tasks   domain.TaskRepository
	workers domain.WorkerRepository
	queue   domain.Queue
}

// New creates a Scheduler backed by the supplied repositories and queue.
func New(
	tasks domain.TaskRepository,
	workers domain.WorkerRepository,
	queue domain.Queue,
) *Scheduler {
	return &Scheduler{tasks: tasks, workers: workers, queue: queue}
}

// Submit validates task, transitions it to Queued, persists it, and enqueues
// it for execution. Returns domain.ErrTaskInvalid (wrapped) if validation fails.
func (s *Scheduler) Submit(ctx context.Context, task *domain.Task) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrTaskInvalid, err)
	}
	now := time.Now()
	task.Status = domain.TaskStatusQueued
	task.UpdatedAt = now
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	if err := s.tasks.Save(ctx, task); err != nil {
		return err
	}
	return s.queue.Enqueue(ctx, task)
}

// Cancel marks the task as Failed if it has not yet reached a terminal state.
// Cancelling an already-terminal task is a no-op.
func (s *Scheduler) Cancel(ctx context.Context, taskID string) error {
	task, err := s.tasks.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task.IsTerminal() {
		return nil
	}
	task.Status = domain.TaskStatusFailed
	task.UpdatedAt = time.Now()
	return s.tasks.Save(ctx, task)
}

// Status returns the current TaskStatus for the given taskID.
func (s *Scheduler) Status(ctx context.Context, taskID string) (domain.TaskStatus, error) {
	task, err := s.tasks.FindByID(ctx, taskID)
	if err != nil {
		return "", err
	}
	return task.Status, nil
}
