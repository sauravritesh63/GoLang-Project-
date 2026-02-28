// Package scheduler provides the task queue and scheduling logic for the
// distributed task scheduler.
package scheduler

import (
	"context"
	"sync"

	"github.com/sauravritesh63/GoLang-Project-/domain"
)

// MemQueue is a thread-safe, unbounded in-memory implementation of domain.Queue.
// Tasks are served in FIFO order.
type MemQueue struct {
	mu  sync.Mutex
	buf []*domain.Task
	sig chan struct{}
}

// NewMemQueue creates an empty MemQueue ready for use.
func NewMemQueue() *MemQueue {
	return &MemQueue{sig: make(chan struct{}, 1)}
}

// Enqueue appends task to the tail of the queue and notifies any blocked
// Dequeue callers.
func (q *MemQueue) Enqueue(_ context.Context, task *domain.Task) error {
	q.mu.Lock()
	q.buf = append(q.buf, task)
	q.mu.Unlock()
	select {
	case q.sig <- struct{}{}:
	default:
	}
	return nil
}

// Dequeue removes and returns the head task. It blocks until a task is
// available or ctx is cancelled, in which case domain.ErrQueueEmpty is returned.
func (q *MemQueue) Dequeue(ctx context.Context) (*domain.Task, error) {
	for {
		q.mu.Lock()
		if len(q.buf) > 0 {
			t := q.buf[0]
			q.buf = q.buf[1:]
			remaining := len(q.buf)
			q.mu.Unlock()
			// Re-signal so other waiting callers can wake up when tasks remain.
			if remaining > 0 {
				select {
				case q.sig <- struct{}{}:
				default:
				}
			}
			return t, nil
		}
		q.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, domain.ErrQueueEmpty
		case <-q.sig:
		}
	}
}

// Len returns the number of tasks currently waiting in the queue.
func (q *MemQueue) Len(_ context.Context) (int, error) {
	q.mu.Lock()
	n := len(q.buf)
	q.mu.Unlock()
	return n, nil
}
