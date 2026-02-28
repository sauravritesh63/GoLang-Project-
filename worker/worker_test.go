package worker_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sauravritesh63/GoLang-Project-/domain"
	"github.com/sauravritesh63/GoLang-Project-/scheduler"
	"github.com/sauravritesh63/GoLang-Project-/worker"
)

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
		Status:      domain.TaskStatusQueued,
		Priority:    domain.PriorityNormal,
		MaxRetries:  2,
		ScheduledAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// poll calls check every 5 ms until it returns true or the timeout expires.
func poll(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

// ── Worker.Run tests ──────────────────────────────────────────────────────────

func TestWorker_Run_Registers(t *testing.T) {
	q := scheduler.NewMemQueue()
	tr := newMemTaskRepo()
	wr := newMemWorkerRepo()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := worker.New("w1", q, tr, wr, func(_ context.Context, _ *domain.Task) error { return nil })
	errCh := make(chan error, 1)
	go func() { errCh <- w.Run(ctx) }()

	// Poll until the worker has registered itself.
	poll(t, time.Second, func() bool {
		_, err := wr.FindByID(ctx, "w1")
		return err == nil
	})

	wrk, err := wr.FindByID(ctx, "w1")
	if err != nil {
		t.Fatalf("worker not registered: %v", err)
	}
	if wrk.ID != "w1" {
		t.Errorf("worker ID: got %q, want w1", wrk.ID)
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestWorker_Run_SuccessfulTask(t *testing.T) {
	q := scheduler.NewMemQueue()
	tr := newMemTaskRepo()
	wr := newMemWorkerRepo()

	task := validTask("t1")
	_ = tr.Save(context.Background(), task)
	_ = q.Enqueue(context.Background(), task)

	executed := make(chan string, 1)
	h := func(_ context.Context, t *domain.Task) error {
		executed <- t.ID
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	w := worker.New("w1", q, tr, wr, h)
	errCh := make(chan error, 1)
	go func() { errCh <- w.Run(ctx) }()

	select {
	case id := <-executed:
		if id != "t1" {
			t.Errorf("executed task ID: got %q, want t1", id)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for task execution")
	}

	// Poll until the worker has persisted the final status.
	poll(t, time.Second, func() bool {
		stored, _ := tr.FindByID(context.Background(), "t1")
		return stored != nil && stored.Status == domain.TaskStatusSucceeded
	})
	cancel()
	<-errCh

	stored, err := tr.FindByID(context.Background(), "t1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if stored.Status != domain.TaskStatusSucceeded {
		t.Errorf("task status: got %q, want succeeded", stored.Status)
	}
	if stored.StartedAt == nil {
		t.Error("StartedAt should be set after execution")
	}
	if stored.FinishedAt == nil {
		t.Error("FinishedAt should be set after successful execution")
	}
}

func TestWorker_Run_FailedTaskWithRetry(t *testing.T) {
	q := scheduler.NewMemQueue()
	tr := newMemTaskRepo()
	wr := newMemWorkerRepo()

	task := validTask("t1")
	task.MaxRetries = 1
	_ = tr.Save(context.Background(), task)
	_ = q.Enqueue(context.Background(), task)

	attempts := 0
	h := func(_ context.Context, _ *domain.Task) error {
		attempts++
		return errors.New("task failed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	noBackoff := worker.WithBackoff(func(int) time.Duration { return 0 })
	w := worker.New("w1", q, tr, wr, h, noBackoff)
	errCh := make(chan error, 1)
	go func() { errCh <- w.Run(ctx) }()

	// Poll until the task reaches a terminal state.
	poll(t, 2*time.Second, func() bool {
		stored, _ := tr.FindByID(context.Background(), "t1")
		return stored != nil && stored.IsTerminal()
	})

	cancel()
	<-errCh

	stored, err := tr.FindByID(context.Background(), "t1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if stored.Status != domain.TaskStatusFailed {
		t.Errorf("task status: got %q, want failed", stored.Status)
	}
	// MaxRetries=1 means 1 retry → 2 total attempts.
	if attempts != 2 {
		t.Errorf("expected 2 attempts (1 initial + 1 retry), got %d", attempts)
	}
}

func TestWorker_Run_NoRetry_WhenMaxRetriesZero(t *testing.T) {
	q := scheduler.NewMemQueue()
	tr := newMemTaskRepo()
	wr := newMemWorkerRepo()

	task := validTask("t1")
	task.MaxRetries = 0
	_ = tr.Save(context.Background(), task)
	_ = q.Enqueue(context.Background(), task)

	attempts := 0
	h := func(_ context.Context, _ *domain.Task) error {
		attempts++
		return errors.New("immediate failure")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	w := worker.New("w1", q, tr, wr, h)
	errCh := make(chan error, 1)
	go func() { errCh <- w.Run(ctx) }()

	poll(t, 2*time.Second, func() bool {
		stored, _ := tr.FindByID(context.Background(), "t1")
		return stored != nil && stored.IsTerminal()
	})

	cancel()
	<-errCh

	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retries), got %d", attempts)
	}
	stored, _ := tr.FindByID(context.Background(), "t1")
	if stored.Status != domain.TaskStatusFailed {
		t.Errorf("task status: got %q, want failed", stored.Status)
	}
}

func TestWorker_Run_CleanShutdown(t *testing.T) {
	q := scheduler.NewMemQueue()
	tr := newMemTaskRepo()
	wr := newMemWorkerRepo()

	ctx, cancel := context.WithCancel(context.Background())
	h := func(_ context.Context, _ *domain.Task) error { return nil }
	w := worker.New("w1", q, tr, wr, h)

	errCh := make(chan error, 1)
	go func() { errCh <- w.Run(ctx) }()

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("expected nil error on clean shutdown, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not shut down within 1 second")
	}
}

func TestWorker_Run_HeartbeatUpdated(t *testing.T) {
	q := scheduler.NewMemQueue()
	tr := newMemTaskRepo()
	wr := newMemWorkerRepo()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a short heartbeat interval for testing.
	h := func(_ context.Context, _ *domain.Task) error { return nil }
	w := worker.New("w1", q, tr, wr, h,
		worker.WithHeartbeatInterval(50*time.Millisecond),
	)

	go func() { _ = w.Run(ctx) }()

	// Poll until the worker has registered itself.
	poll(t, time.Second, func() bool {
		_, err := wr.FindByID(ctx, "w1")
		return err == nil
	})

	before, err := wr.FindByID(ctx, "w1")
	if err != nil {
		t.Fatalf("FindByID before heartbeat: %v", err)
	}
	initialHeart := before.LastHeartAt

	// Poll until the heartbeat is updated.
	poll(t, time.Second, func() bool {
		w, _ := wr.FindByID(ctx, "w1")
		return w != nil && w.LastHeartAt.After(initialHeart)
	})

	after, err := wr.FindByID(ctx, "w1")
	if err != nil {
		t.Fatalf("FindByID after heartbeat: %v", err)
	}
	if !after.LastHeartAt.After(initialHeart) {
		t.Errorf("expected LastHeartAt to be updated; before=%v after=%v", initialHeart, after.LastHeartAt)
	}
}

func TestMockShellHandler(t *testing.T) {
	ctx := context.Background()

	// Task with payload.
	task := validTask("t1")
	task.Payload = []byte("echo hello")
	if err := worker.MockShellHandler(ctx, task); err != nil {
		t.Errorf("MockShellHandler with payload: unexpected error: %v", err)
	}

	// Task without payload.
	empty := validTask("t2")
	if err := worker.MockShellHandler(ctx, empty); err != nil {
		t.Errorf("MockShellHandler without payload: unexpected error: %v", err)
	}
}

func TestWorker_ExponentialBackoff(t *testing.T) {
	// Verify that the backoff delay is applied between retries.
	q := scheduler.NewMemQueue()
	tr := newMemTaskRepo()
	wr := newMemWorkerRepo()

	task := validTask("t1")
	task.MaxRetries = 1
	_ = tr.Save(context.Background(), task)
	_ = q.Enqueue(context.Background(), task)

	// Record the timestamps of each attempt.
	var (
		mu         sync.Mutex
		timestamps []time.Time
	)
	h := func(_ context.Context, _ *domain.Task) error {
		mu.Lock()
		timestamps = append(timestamps, time.Now())
		mu.Unlock()
		return errors.New("always fail")
	}

	const backoffDelay = 50 * time.Millisecond
	fixedBackoff := worker.WithBackoff(func(int) time.Duration { return backoffDelay })

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	w := worker.New("w1", q, tr, wr, h, fixedBackoff)
	errCh := make(chan error, 1)
	go func() { errCh <- w.Run(ctx) }()

	// Wait for terminal state (2 attempts with MaxRetries=1).
	poll(t, 2*time.Second, func() bool {
		stored, _ := tr.FindByID(context.Background(), "t1")
		return stored != nil && stored.IsTerminal()
	})
	cancel()
	<-errCh

	mu.Lock()
	defer mu.Unlock()
	if len(timestamps) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(timestamps))
	}
	gap := timestamps[1].Sub(timestamps[0])
	if gap < backoffDelay {
		t.Errorf("expected backoff delay ≥ %v between retries, got %v", backoffDelay, gap)
	}
}
