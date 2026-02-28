package domain_test

import (
	"testing"
	"time"

	"github.com/sauravritesh63/GoLang-Project-/domain"
)

// ── Task tests ────────────────────────────────────────────────────────────────

func validTask() *domain.Task {
	return &domain.Task{
		ID:          "task-1",
		Name:        "send-email",
		Payload:     []byte(`{"to":"user@example.com"}`),
		Status:      domain.TaskStatusPending,
		Priority:    domain.PriorityNormal,
		MaxRetries:  3,
		ScheduledAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestTask_Validate_Valid(t *testing.T) {
	task := validTask()
	if err := task.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestTask_Validate_MissingID(t *testing.T) {
	task := validTask()
	task.ID = ""
	if err := task.Validate(); err == nil {
		t.Fatal("expected error for missing ID, got nil")
	}
}

func TestTask_Validate_MissingName(t *testing.T) {
	task := validTask()
	task.Name = ""
	if err := task.Validate(); err == nil {
		t.Fatal("expected error for missing Name, got nil")
	}
}

func TestTask_Validate_InvalidPriorityLow(t *testing.T) {
	task := validTask()
	task.Priority = 0
	if err := task.Validate(); err == nil {
		t.Fatal("expected error for priority 0, got nil")
	}
}

func TestTask_Validate_InvalidPriorityHigh(t *testing.T) {
	task := validTask()
	task.Priority = 11
	if err := task.Validate(); err == nil {
		t.Fatal("expected error for priority 11, got nil")
	}
}

func TestTask_Validate_NegativeMaxRetries(t *testing.T) {
	task := validTask()
	task.MaxRetries = -1
	if err := task.Validate(); err == nil {
		t.Fatal("expected error for negative MaxRetries, got nil")
	}
}

func TestTask_CanRetry(t *testing.T) {
	task := validTask()
	task.MaxRetries = 3
	task.RetryCount = 2
	if !task.CanRetry() {
		t.Fatal("expected CanRetry to return true")
	}
	task.RetryCount = 3
	if task.CanRetry() {
		t.Fatal("expected CanRetry to return false when RetryCount == MaxRetries")
	}
}

func TestTask_IsTerminal(t *testing.T) {
	cases := []struct {
		status   domain.TaskStatus
		terminal bool
	}{
		{domain.TaskStatusPending, false},
		{domain.TaskStatusQueued, false},
		{domain.TaskStatusRunning, false},
		{domain.TaskStatusRetrying, false},
		{domain.TaskStatusSucceeded, true},
		{domain.TaskStatusFailed, true},
	}
	for _, tc := range cases {
		task := validTask()
		task.Status = tc.status
		if got := task.IsTerminal(); got != tc.terminal {
			t.Errorf("IsTerminal(%s) = %v, want %v", tc.status, got, tc.terminal)
		}
	}
}

// ── Worker tests ──────────────────────────────────────────────────────────────

func validWorker() *domain.Worker {
	return &domain.Worker{
		ID:           "worker-1",
		Address:      "127.0.0.1:9090",
		Status:       domain.WorkerStatusIdle,
		Concurrency:  4,
		ActiveTasks:  0,
		LastHeartAt:  time.Now(),
		RegisteredAt: time.Now(),
	}
}

func TestWorker_Validate_Valid(t *testing.T) {
	w := validWorker()
	if err := w.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestWorker_Validate_MissingID(t *testing.T) {
	w := validWorker()
	w.ID = ""
	if err := w.Validate(); err == nil {
		t.Fatal("expected error for missing ID, got nil")
	}
}

func TestWorker_Validate_MissingAddress(t *testing.T) {
	w := validWorker()
	w.Address = ""
	if err := w.Validate(); err == nil {
		t.Fatal("expected error for missing Address, got nil")
	}
}

func TestWorker_Validate_ZeroConcurrency(t *testing.T) {
	w := validWorker()
	w.Concurrency = 0
	if err := w.Validate(); err == nil {
		t.Fatal("expected error for zero Concurrency, got nil")
	}
}

func TestWorker_HasCapacity_Idle(t *testing.T) {
	w := validWorker()
	w.Status = domain.WorkerStatusIdle
	if !w.HasCapacity() {
		t.Fatal("expected idle worker to have capacity")
	}
}

func TestWorker_HasCapacity_BusyWithRoom(t *testing.T) {
	w := validWorker()
	w.Status = domain.WorkerStatusBusy
	w.Concurrency = 4
	w.ActiveTasks = 2
	if !w.HasCapacity() {
		t.Fatal("expected busy worker with free slots to have capacity")
	}
}

func TestWorker_HasCapacity_BusyAtLimit(t *testing.T) {
	w := validWorker()
	w.Status = domain.WorkerStatusBusy
	w.Concurrency = 4
	w.ActiveTasks = 4
	if w.HasCapacity() {
		t.Fatal("expected busy worker at limit to have no capacity")
	}
}

func TestWorker_HasCapacity_Offline(t *testing.T) {
	w := validWorker()
	w.Status = domain.WorkerStatusOffline
	if w.HasCapacity() {
		t.Fatal("expected offline worker to have no capacity")
	}
}

func TestWorker_IsAlive(t *testing.T) {
	w := validWorker()
	w.LastHeartAt = time.Now()
	if !w.IsAlive(5 * time.Second) {
		t.Fatal("expected freshly-heartbeated worker to be alive")
	}

	w.LastHeartAt = time.Now().Add(-10 * time.Second)
	if w.IsAlive(5 * time.Second) {
		t.Fatal("expected worker with stale heartbeat to not be alive")
	}
}

// ── Sentinel error tests ──────────────────────────────────────────────────────

func TestSentinelErrors_NotNil(t *testing.T) {
	errs := []error{
		domain.ErrTaskNotFound,
		domain.ErrWorkerNotFound,
		domain.ErrQueueEmpty,
		domain.ErrTaskInvalid,
		domain.ErrWorkerInvalid,
	}
	for _, err := range errs {
		if err == nil {
			t.Errorf("expected non-nil sentinel error, got nil")
		}
	}
}
