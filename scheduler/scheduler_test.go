package scheduler_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sauravritesh63/GoLang-Project-/domain"
	"github.com/sauravritesh63/GoLang-Project-/scheduler"
)

var ctx = context.Background()

// ── in-memory repositories ────────────────────────────────────────────────────

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

// ── helpers ───────────────────────────────────────────────────────────────────

func validTask(id string) *domain.Task {
	return &domain.Task{
		ID:          id,
		Name:        "send-email",
		Status:      domain.TaskStatusPending,
		Priority:    domain.PriorityNormal,
		MaxRetries:  2,
		ScheduledAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func newScheduler() (*scheduler.Scheduler, *memTaskRepo) {
	tr := newMemTaskRepo()
	wr := newMemWorkerRepo()
	q := scheduler.NewMemQueue()
	return scheduler.New(tr, wr, q), tr
}

// ── MemQueue tests ────────────────────────────────────────────────────────────

func TestMemQueue_EnqueueDequeue(t *testing.T) {
	q := scheduler.NewMemQueue()
	task := validTask("t1")
	if err := q.Enqueue(ctx, task); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	got, err := q.Dequeue(ctx)
	if err != nil {
		t.Fatalf("Dequeue: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("got task ID %q, want %q", got.ID, task.ID)
	}
}

func TestMemQueue_FIFOOrder(t *testing.T) {
	q := scheduler.NewMemQueue()
	_ = q.Enqueue(ctx, validTask("t1"))
	_ = q.Enqueue(ctx, validTask("t2"))
	_ = q.Enqueue(ctx, validTask("t3"))

	for _, want := range []string{"t1", "t2", "t3"} {
		got, err := q.Dequeue(ctx)
		if err != nil {
			t.Fatalf("Dequeue: %v", err)
		}
		if got.ID != want {
			t.Errorf("FIFO order: got %q, want %q", got.ID, want)
		}
	}
}

func TestMemQueue_Len(t *testing.T) {
	q := scheduler.NewMemQueue()
	_ = q.Enqueue(ctx, validTask("t1"))
	_ = q.Enqueue(ctx, validTask("t2"))
	n, err := q.Len(ctx)
	if err != nil {
		t.Fatalf("Len: %v", err)
	}
	if n != 2 {
		t.Errorf("Len: got %d, want 2", n)
	}
}

func TestMemQueue_Len_Empty(t *testing.T) {
	q := scheduler.NewMemQueue()
	n, err := q.Len(ctx)
	if err != nil {
		t.Fatalf("Len: %v", err)
	}
	if n != 0 {
		t.Errorf("Len on empty queue: got %d, want 0", n)
	}
}

func TestMemQueue_Dequeue_ContextCancelled(t *testing.T) {
	q := scheduler.NewMemQueue()
	ctx2, cancel := context.WithCancel(ctx)
	cancel() // already cancelled
	_, err := q.Dequeue(ctx2)
	if err == nil {
		t.Fatal("expected error on cancelled context, got nil")
	}
	if !errors.Is(err, domain.ErrQueueEmpty) {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}
}

// ── Scheduler.Submit tests ────────────────────────────────────────────────────

func TestScheduler_Submit_Valid(t *testing.T) {
	sched, repo := newScheduler()
	task := validTask("t1")
	if err := sched.Submit(ctx, task); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	stored, err := repo.FindByID(ctx, "t1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if stored.Status != domain.TaskStatusQueued {
		t.Errorf("Status: got %q, want %q", stored.Status, domain.TaskStatusQueued)
	}
}

func TestScheduler_Submit_InvalidTask_MissingID(t *testing.T) {
	sched, _ := newScheduler()
	task := validTask("")
	err := sched.Submit(ctx, task)
	if err == nil {
		t.Fatal("expected error for task with empty ID, got nil")
	}
	if !errors.Is(err, domain.ErrTaskInvalid) {
		t.Errorf("expected ErrTaskInvalid, got %v", err)
	}
}

func TestScheduler_Submit_EnqueuesTask(t *testing.T) {
	tr := newMemTaskRepo()
	wr := newMemWorkerRepo()
	q := scheduler.NewMemQueue()
	sched := scheduler.New(tr, wr, q)
	task := validTask("t1")
	if err := sched.Submit(ctx, task); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	n, _ := q.Len(ctx)
	if n != 1 {
		t.Errorf("queue length after Submit: got %d, want 1", n)
	}
}

func TestScheduler_Submit_SetsCreatedAt(t *testing.T) {
	sched, repo := newScheduler()
	task := validTask("t1")
	task.CreatedAt = time.Time{} // zero value
	before := time.Now()
	_ = sched.Submit(ctx, task)
	stored, _ := repo.FindByID(ctx, "t1")
	if stored.CreatedAt.Before(before) {
		t.Error("expected CreatedAt to be set on Submit")
	}
}

// ── Scheduler.Cancel tests ────────────────────────────────────────────────────

func TestScheduler_Cancel_QueuedTask(t *testing.T) {
	sched, repo := newScheduler()
	task := validTask("t1")
	_ = sched.Submit(ctx, task)
	if err := sched.Cancel(ctx, "t1"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	stored, _ := repo.FindByID(ctx, "t1")
	if stored.Status != domain.TaskStatusFailed {
		t.Errorf("Status after Cancel: got %q, want %q", stored.Status, domain.TaskStatusFailed)
	}
}

func TestScheduler_Cancel_TerminalTask_NoOp(t *testing.T) {
	sched, repo := newScheduler()
	task := validTask("t1")
	_ = sched.Submit(ctx, task)
	// Manually mark as succeeded.
	stored, _ := repo.FindByID(ctx, "t1")
	stored.Status = domain.TaskStatusSucceeded
	_ = repo.Save(ctx, stored)
	// Cancel of a terminal task should be a no-op.
	if err := sched.Cancel(ctx, "t1"); err != nil {
		t.Fatalf("Cancel on terminal task: %v", err)
	}
	got, _ := repo.FindByID(ctx, "t1")
	if got.Status != domain.TaskStatusSucceeded {
		t.Errorf("expected status to remain succeeded, got %q", got.Status)
	}
}

func TestScheduler_Cancel_NotFound(t *testing.T) {
	sched, _ := newScheduler()
	err := sched.Cancel(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown task, got nil")
	}
}

// ── Scheduler.Status tests ────────────────────────────────────────────────────

func TestScheduler_Status_Queued(t *testing.T) {
	sched, _ := newScheduler()
	task := validTask("t1")
	_ = sched.Submit(ctx, task)
	status, err := sched.Status(ctx, "t1")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status != domain.TaskStatusQueued {
		t.Errorf("Status: got %q, want %q", status, domain.TaskStatusQueued)
	}
}

func TestScheduler_Status_NotFound(t *testing.T) {
	sched, _ := newScheduler()
	_, err := sched.Status(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown task, got nil")
	}
}

// ── interface compliance ──────────────────────────────────────────────────────

var (
	_ domain.Queue     = (*scheduler.MemQueue)(nil)
	_ domain.Scheduler = (*scheduler.Scheduler)(nil)
)
