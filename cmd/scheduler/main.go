// Package main is the entry point for the distributed task scheduler service.
// It wires up the in-memory queue, task and worker repositories, and waits for
// shutdown. In production, replace the in-memory stores with Redis / Postgres
// backed implementations via environment variables.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sauravritesh63/GoLang-Project-/domain"
	"github.com/sauravritesh63/GoLang-Project-/scheduler"
)

func main() {
	queue := scheduler.NewMemQueue()
	taskRepo := newMemTaskRepo()
	workerRepo := newMemWorkerRepo()

	// In production, the scheduler would be invoked via an RPC/HTTP endpoint
	// or a cron-triggered goroutine. Here we just wire it up to confirm
	// dependencies resolve correctly at startup.
	sched := scheduler.New(taskRepo, workerRepo, queue)
	log.Printf("Scheduler initialised (queue depth: %T)", sched)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("Scheduler service started; waiting for shutdown signal")
	<-ctx.Done()
	log.Println("Scheduler service stopped")
}

// ── in-memory stores (replace with Redis/Postgres in production) ──────────────

type memTaskRepo struct {
	mu    sync.RWMutex
	store map[string]*domain.Task
}

func newMemTaskRepo() *memTaskRepo {
	return &memTaskRepo{store: make(map[string]*domain.Task)}
}

func (r *memTaskRepo) Save(_ context.Context, t *domain.Task) error {
	r.mu.Lock()
	cp := *t
	r.store[t.ID] = &cp
	r.mu.Unlock()
	return nil
}

func (r *memTaskRepo) FindByID(_ context.Context, id string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.store[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *memTaskRepo) FindByStatus(_ context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Task
	for _, t := range r.store {
		if t.Status == status {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memTaskRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return domain.ErrTaskNotFound
	}
	delete(r.store, id)
	return nil
}

type memWorkerRepo struct {
	mu    sync.RWMutex
	store map[string]*domain.Worker
}

func newMemWorkerRepo() *memWorkerRepo {
	return &memWorkerRepo{store: make(map[string]*domain.Worker)}
}

func (r *memWorkerRepo) Save(_ context.Context, w *domain.Worker) error {
	r.mu.Lock()
	cp := *w
	r.store[w.ID] = &cp
	r.mu.Unlock()
	return nil
}

func (r *memWorkerRepo) FindByID(_ context.Context, id string) (*domain.Worker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.store[id]
	if !ok {
		return nil, domain.ErrWorkerNotFound
	}
	cp := *w
	return &cp, nil
}

func (r *memWorkerRepo) FindAvailable(_ context.Context) ([]*domain.Worker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Worker
	for _, w := range r.store {
		if w.HasCapacity() {
			cp := *w
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memWorkerRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return domain.ErrWorkerNotFound
	}
	delete(r.store, id)
	return nil
}
