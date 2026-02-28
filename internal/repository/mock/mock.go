// Package mock provides in-memory implementations of the repository interfaces
// defined in the parent package. They are intended for unit testing and are
// safe for concurrent use (protected by a sync.RWMutex per store).
package mock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
)

// ── WorkflowRepository ────────────────────────────────────────────────────────

// WorkflowRepo is an in-memory WorkflowRepository for testing.
type WorkflowRepo struct {
	mu    sync.RWMutex
	store map[uuid.UUID]*domain.Workflow
}

// NewWorkflowRepo returns an empty in-memory WorkflowRepo.
func NewWorkflowRepo() *WorkflowRepo {
	return &WorkflowRepo{store: make(map[uuid.UUID]*domain.Workflow)}
}

func (r *WorkflowRepo) Create(_ context.Context, wf *domain.Workflow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *wf
	r.store[wf.ID] = &cp
	return nil
}

func (r *WorkflowRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Workflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wf, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *wf
	return &cp, nil
}

func (r *WorkflowRepo) Update(_ context.Context, wf *domain.Workflow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[wf.ID]; !ok {
		return repository.ErrNotFound
	}
	cp := *wf
	r.store[wf.ID] = &cp
	return nil
}

func (r *WorkflowRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.store, id)
	return nil
}

func (r *WorkflowRepo) List(_ context.Context) ([]*domain.Workflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Workflow, 0, len(r.store))
	for _, wf := range r.store {
		cp := *wf
		out = append(out, &cp)
	}
	return out, nil
}

func (r *WorkflowRepo) ListActive(_ context.Context) ([]*domain.Workflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Workflow
	for _, wf := range r.store {
		if wf.IsActive {
			cp := *wf
			out = append(out, &cp)
		}
	}
	return out, nil
}

// ── TaskRepository ────────────────────────────────────────────────────────────

// TaskRepo is an in-memory TaskRepository for testing.
type TaskRepo struct {
	mu    sync.RWMutex
	store map[uuid.UUID]*domain.Task
}

// NewTaskRepo returns an empty in-memory TaskRepo.
func NewTaskRepo() *TaskRepo {
	return &TaskRepo{store: make(map[uuid.UUID]*domain.Task)}
}

func (r *TaskRepo) Create(_ context.Context, t *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *t
	r.store[t.ID] = &cp
	return nil
}

func (r *TaskRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *TaskRepo) Update(_ context.Context, t *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[t.ID]; !ok {
		return repository.ErrNotFound
	}
	cp := *t
	r.store[t.ID] = &cp
	return nil
}

func (r *TaskRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.store, id)
	return nil
}

func (r *TaskRepo) ListByWorkflowID(_ context.Context, workflowID uuid.UUID) ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Task
	for _, t := range r.store {
		if t.WorkflowID == workflowID {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out, nil
}

// ── WorkflowRunRepository ─────────────────────────────────────────────────────

// WorkflowRunRepo is an in-memory WorkflowRunRepository for testing.
type WorkflowRunRepo struct {
	mu    sync.RWMutex
	store map[uuid.UUID]*domain.WorkflowRun
}

// NewWorkflowRunRepo returns an empty in-memory WorkflowRunRepo.
func NewWorkflowRunRepo() *WorkflowRunRepo {
	return &WorkflowRunRepo{store: make(map[uuid.UUID]*domain.WorkflowRun)}
}

func (r *WorkflowRunRepo) Create(_ context.Context, wr *domain.WorkflowRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *wr
	r.store[wr.ID] = &cp
	return nil
}

func (r *WorkflowRunRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.WorkflowRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wr, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *wr
	return &cp, nil
}

func (r *WorkflowRunRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.Status, finishedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	wr, ok := r.store[id]
	if !ok {
		return repository.ErrNotFound
	}
	wr.Status = status
	wr.FinishedAt = finishedAt
	return nil
}

func (r *WorkflowRunRepo) ListByWorkflowID(_ context.Context, workflowID uuid.UUID) ([]*domain.WorkflowRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.WorkflowRun
	for _, wr := range r.store {
		if wr.WorkflowID == workflowID {
			cp := *wr
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *WorkflowRunRepo) ListByStatus(_ context.Context, status domain.Status) ([]*domain.WorkflowRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.WorkflowRun
	for _, wr := range r.store {
		if wr.Status == status {
			cp := *wr
			out = append(out, &cp)
		}
	}
	return out, nil
}

// ── TaskRunRepository ─────────────────────────────────────────────────────────

// TaskRunRepo is an in-memory TaskRunRepository for testing.
type TaskRunRepo struct {
	mu    sync.RWMutex
	store map[uuid.UUID]*domain.TaskRun
}

// NewTaskRunRepo returns an empty in-memory TaskRunRepo.
func NewTaskRunRepo() *TaskRunRepo {
	return &TaskRunRepo{store: make(map[uuid.UUID]*domain.TaskRun)}
}

func (r *TaskRunRepo) Create(_ context.Context, tr *domain.TaskRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *tr
	r.store[tr.ID] = &cp
	return nil
}

func (r *TaskRunRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.TaskRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tr, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *tr
	return &cp, nil
}

func (r *TaskRunRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.Status, finishedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tr, ok := r.store[id]
	if !ok {
		return repository.ErrNotFound
	}
	tr.Status = status
	tr.FinishedAt = finishedAt
	return nil
}

func (r *TaskRunRepo) ListByWorkflowRunID(_ context.Context, workflowRunID uuid.UUID) ([]*domain.TaskRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.TaskRun
	for _, tr := range r.store {
		if tr.WorkflowRunID == workflowRunID {
			cp := *tr
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *TaskRunRepo) ListByTaskID(_ context.Context, taskID uuid.UUID) ([]*domain.TaskRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.TaskRun
	for _, tr := range r.store {
		if tr.TaskID == taskID {
			cp := *tr
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *TaskRunRepo) ListByStatus(_ context.Context, status domain.Status) ([]*domain.TaskRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.TaskRun
	for _, tr := range r.store {
		if tr.Status == status {
			cp := *tr
			out = append(out, &cp)
		}
	}
	return out, nil
}

// ── WorkerRepository ──────────────────────────────────────────────────────────

// WorkerRepo is an in-memory WorkerRepository for testing.
type WorkerRepo struct {
	mu    sync.RWMutex
	store map[uuid.UUID]*domain.Worker
}

// NewWorkerRepo returns an empty in-memory WorkerRepo.
func NewWorkerRepo() *WorkerRepo {
	return &WorkerRepo{store: make(map[uuid.UUID]*domain.Worker)}
}

func (r *WorkerRepo) Create(_ context.Context, w *domain.Worker) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *w
	r.store[w.ID] = &cp
	return nil
}

func (r *WorkerRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Worker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := *w
	return &cp, nil
}

func (r *WorkerRepo) Update(_ context.Context, w *domain.Worker) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[w.ID]; !ok {
		return repository.ErrNotFound
	}
	cp := *w
	r.store[w.ID] = &cp
	return nil
}

func (r *WorkerRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.store, id)
	return nil
}

func (r *WorkerRepo) ListActive(_ context.Context) ([]*domain.Worker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Worker
	for _, w := range r.store {
		if w.Status == domain.WorkerStatusActive {
			cp := *w
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *WorkerRepo) UpdateHeartbeat(_ context.Context, id uuid.UUID, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	w, ok := r.store[id]
	if !ok {
		return repository.ErrNotFound
	}
	w.LastHeartbeat = at
	return nil
}
