// Package worker provides the worker service for the distributed task scheduler.
// A Worker pulls tasks from the Queue, executes them via a Handler function,
// and persists status updates through the domain repositories.
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/sauravritesh63/GoLang-Project-/domain"
)

// Handler is the function type responsible for executing a task's payload.
// It should return nil on success or a non-nil error on failure.
type Handler func(ctx context.Context, task *domain.Task) error

// Worker dequeues tasks from a Queue, executes them using a Handler, and
// manages task lifecycle: status transitions, retries, and heartbeats.
type Worker struct {
	id      string
	queue   domain.Queue
	tasks   domain.TaskRepository
	workers domain.WorkerRepository
	handler Handler

	heartbeatInterval time.Duration
}

// Option is a functional option for configuring a Worker.
type Option func(*Worker)

// WithHeartbeatInterval sets the interval between heartbeat updates.
// The default is 15 seconds.
func WithHeartbeatInterval(d time.Duration) Option {
	return func(w *Worker) { w.heartbeatInterval = d }
}

// New creates a Worker with the given ID, dependencies, and task handler.
func New(
	id string,
	queue domain.Queue,
	tasks domain.TaskRepository,
	workers domain.WorkerRepository,
	handler Handler,
	opts ...Option,
) *Worker {
	w := &Worker{
		id:                id,
		queue:             queue,
		tasks:             tasks,
		workers:           workers,
		handler:           handler,
		heartbeatInterval: 15 * time.Second,
	}
	for _, o := range opts {
		o(w)
	}
	return w
}

// Run registers the worker, starts the heartbeat loop, and processes tasks
// until ctx is cancelled. It always returns nil when the context expires.
func (w *Worker) Run(ctx context.Context) error {
	now := time.Now()
	wrk := &domain.Worker{
		ID:           w.id,
		Address:      w.id,
		Status:       domain.WorkerStatusIdle,
		Concurrency:  1,
		ActiveTasks:  0,
		LastHeartAt:  now,
		RegisteredAt: now,
	}
	if err := w.workers.Save(ctx, wrk); err != nil {
		return fmt.Errorf("worker register: %w", err)
	}

	go w.heartbeatLoop(ctx)

	for {
		task, err := w.queue.Dequeue(ctx)
		if err != nil {
			// Context cancelled — clean shutdown.
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		w.execute(ctx, task)
	}
}

// execute runs a single task, handling status transitions and retry logic.
func (w *Worker) execute(ctx context.Context, task *domain.Task) {
	now := time.Now()
	task.Status = domain.TaskStatusRunning
	task.StartedAt = &now
	task.UpdatedAt = now
	_ = w.tasks.Save(ctx, task)

	err := w.handler(ctx, task)

	finished := time.Now()
	task.UpdatedAt = finished

	if err == nil {
		task.FinishedAt = &finished
		task.Status = domain.TaskStatusSucceeded
		task.Error = ""
	} else {
		task.Error = err.Error()
		if task.CanRetry() {
			task.RetryCount++
			task.Status = domain.TaskStatusRetrying
			_ = w.tasks.Save(ctx, task)
			// Re-enqueue for retry.
			_ = w.queue.Enqueue(ctx, task)
			return
		}
		task.FinishedAt = &finished
		task.Status = domain.TaskStatusFailed
	}
	_ = w.tasks.Save(ctx, task)
}

// heartbeatLoop updates the worker's LastHeartAt at the configured interval
// until ctx is cancelled.
func (w *Worker) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(w.heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			wrk, err := w.workers.FindByID(ctx, w.id)
			if err != nil {
				continue
			}
			wrk.LastHeartAt = time.Now()
			_ = w.workers.Save(ctx, wrk)
		}
	}
}
