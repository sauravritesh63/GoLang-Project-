package mock_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository/mock"
)

var ctx = context.Background()

// ── helpers ───────────────────────────────────────────────────────────────────

func newWorkflow() *domain.Workflow {
	return &domain.Workflow{
		ID:           uuid.New(),
		Name:         "etl",
		Description:  "Daily ETL",
		ScheduleCron: "0 2 * * *",
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
	}
}

func newTask(workflowID uuid.UUID) *domain.Task {
	return &domain.Task{
		ID:         uuid.New(),
		WorkflowID: workflowID,
		Name:       "extract",
		Command:    "python extract.py",
		CreatedAt:  time.Now().UTC(),
	}
}

func newWorkflowRun(workflowID uuid.UUID) *domain.WorkflowRun {
	return &domain.WorkflowRun{
		ID:         uuid.New(),
		WorkflowID: workflowID,
		Status:     domain.StatusPending,
		StartedAt:  time.Now().UTC(),
	}
}

func newTaskRun(workflowRunID, taskID uuid.UUID) *domain.TaskRun {
	return &domain.TaskRun{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		TaskID:        taskID,
		Status:        domain.StatusPending,
		Attempt:       1,
		StartedAt:     time.Now().UTC(),
	}
}

func newWorker() *domain.Worker {
	return &domain.Worker{
		ID:            uuid.New(),
		Hostname:      "host-1",
		LastHeartbeat: time.Now().UTC(),
		Status:        domain.WorkerStatusActive,
	}
}

// ── WorkflowRepo ──────────────────────────────────────────────────────────────

func TestWorkflowRepo_CreateAndGetByID(t *testing.T) {
	r := mock.NewWorkflowRepo()
	wf := newWorkflow()

	if err := r.Create(ctx, wf); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := r.GetByID(ctx, wf.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != wf.Name {
		t.Errorf("Name: got %q, want %q", got.Name, wf.Name)
	}
}

func TestWorkflowRepo_GetByID_NotFound(t *testing.T) {
	r := mock.NewWorkflowRepo()
	_, err := r.GetByID(ctx, uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestWorkflowRepo_Update(t *testing.T) {
	r := mock.NewWorkflowRepo()
	wf := newWorkflow()
	_ = r.Create(ctx, wf)

	wf.Name = "updated"
	if err := r.Update(ctx, wf); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := r.GetByID(ctx, wf.ID)
	if got.Name != "updated" {
		t.Errorf("Name after Update: got %q, want %q", got.Name, "updated")
	}
}

func TestWorkflowRepo_Update_NotFound(t *testing.T) {
	r := mock.NewWorkflowRepo()
	err := r.Update(ctx, newWorkflow())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestWorkflowRepo_Delete(t *testing.T) {
	r := mock.NewWorkflowRepo()
	wf := newWorkflow()
	_ = r.Create(ctx, wf)

	if err := r.Delete(ctx, wf.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := r.GetByID(ctx, wf.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound after Delete, got %v", err)
	}
}

func TestWorkflowRepo_Delete_NotFound(t *testing.T) {
	r := mock.NewWorkflowRepo()
	if !errors.Is(r.Delete(ctx, uuid.New()), repository.ErrNotFound) {
		t.Error("expected ErrNotFound")
	}
}

func TestWorkflowRepo_List(t *testing.T) {
	r := mock.NewWorkflowRepo()
	_ = r.Create(ctx, newWorkflow())
	_ = r.Create(ctx, newWorkflow())

	list, err := r.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List length: got %d, want 2", len(list))
	}
}

func TestWorkflowRepo_ListActive(t *testing.T) {
	r := mock.NewWorkflowRepo()
	active := newWorkflow()
	inactive := newWorkflow()
	inactive.IsActive = false
	_ = r.Create(ctx, active)
	_ = r.Create(ctx, inactive)

	list, err := r.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListActive length: got %d, want 1", len(list))
	}
	if list[0].ID != active.ID {
		t.Errorf("ListActive returned wrong workflow")
	}
}

// ── TaskRepo ──────────────────────────────────────────────────────────────────

func TestTaskRepo_CreateAndGetByID(t *testing.T) {
	r := mock.NewTaskRepo()
	tk := newTask(uuid.New())

	if err := r.Create(ctx, tk); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := r.GetByID(ctx, tk.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != tk.Name {
		t.Errorf("Name: got %q, want %q", got.Name, tk.Name)
	}
}

func TestTaskRepo_GetByID_NotFound(t *testing.T) {
	r := mock.NewTaskRepo()
	_, err := r.GetByID(ctx, uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskRepo_ListByWorkflowID(t *testing.T) {
	r := mock.NewTaskRepo()
	wfID := uuid.New()
	_ = r.Create(ctx, newTask(wfID))
	_ = r.Create(ctx, newTask(wfID))
	_ = r.Create(ctx, newTask(uuid.New())) // different workflow

	list, err := r.ListByWorkflowID(ctx, wfID)
	if err != nil {
		t.Fatalf("ListByWorkflowID: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("ListByWorkflowID length: got %d, want 2", len(list))
	}
}

func TestTaskRepo_Delete(t *testing.T) {
	r := mock.NewTaskRepo()
	tk := newTask(uuid.New())
	_ = r.Create(ctx, tk)

	if err := r.Delete(ctx, tk.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !errors.Is(r.Delete(ctx, tk.ID), repository.ErrNotFound) {
		t.Error("expected ErrNotFound on second Delete")
	}
}

// ── WorkflowRunRepo ───────────────────────────────────────────────────────────

func TestWorkflowRunRepo_CreateAndGetByID(t *testing.T) {
	r := mock.NewWorkflowRunRepo()
	wr := newWorkflowRun(uuid.New())

	if err := r.Create(ctx, wr); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := r.GetByID(ctx, wr.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Status != domain.StatusPending {
		t.Errorf("Status: got %q, want pending", got.Status)
	}
}

func TestWorkflowRunRepo_UpdateStatus(t *testing.T) {
	r := mock.NewWorkflowRunRepo()
	wr := newWorkflowRun(uuid.New())
	_ = r.Create(ctx, wr)

	now := time.Now().UTC()
	if err := r.UpdateStatus(ctx, wr.ID, domain.StatusSuccess, &now); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ := r.GetByID(ctx, wr.ID)
	if got.Status != domain.StatusSuccess {
		t.Errorf("Status after UpdateStatus: got %q, want success", got.Status)
	}
	if got.FinishedAt == nil {
		t.Error("FinishedAt should not be nil after UpdateStatus")
	}
}

func TestWorkflowRunRepo_UpdateStatus_NotFound(t *testing.T) {
	r := mock.NewWorkflowRunRepo()
	err := r.UpdateStatus(ctx, uuid.New(), domain.StatusFailed, nil)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestWorkflowRunRepo_ListByWorkflowID(t *testing.T) {
	r := mock.NewWorkflowRunRepo()
	wfID := uuid.New()
	_ = r.Create(ctx, newWorkflowRun(wfID))
	_ = r.Create(ctx, newWorkflowRun(wfID))
	_ = r.Create(ctx, newWorkflowRun(uuid.New()))

	list, err := r.ListByWorkflowID(ctx, wfID)
	if err != nil {
		t.Fatalf("ListByWorkflowID: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("ListByWorkflowID length: got %d, want 2", len(list))
	}
}

func TestWorkflowRunRepo_ListByStatus(t *testing.T) {
	r := mock.NewWorkflowRunRepo()
	_ = r.Create(ctx, newWorkflowRun(uuid.New())) // pending
	run2 := newWorkflowRun(uuid.New())
	run2.Status = domain.StatusRunning
	_ = r.Create(ctx, run2)

	list, err := r.ListByStatus(ctx, domain.StatusPending)
	if err != nil {
		t.Fatalf("ListByStatus: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListByStatus(pending) length: got %d, want 1", len(list))
	}
}

// ── TaskRunRepo ───────────────────────────────────────────────────────────────

func TestTaskRunRepo_CreateAndGetByID(t *testing.T) {
	r := mock.NewTaskRunRepo()
	tr := newTaskRun(uuid.New(), uuid.New())

	if err := r.Create(ctx, tr); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := r.GetByID(ctx, tr.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Attempt != 1 {
		t.Errorf("Attempt: got %d, want 1", got.Attempt)
	}
}

func TestTaskRunRepo_UpdateStatus(t *testing.T) {
	r := mock.NewTaskRunRepo()
	tr := newTaskRun(uuid.New(), uuid.New())
	_ = r.Create(ctx, tr)

	if err := r.UpdateStatus(ctx, tr.ID, domain.StatusSuccess, nil); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ := r.GetByID(ctx, tr.ID)
	if got.Status != domain.StatusSuccess {
		t.Errorf("Status: got %q, want success", got.Status)
	}
}

func TestTaskRunRepo_ListByWorkflowRunID(t *testing.T) {
	r := mock.NewTaskRunRepo()
	wrID := uuid.New()
	_ = r.Create(ctx, newTaskRun(wrID, uuid.New()))
	_ = r.Create(ctx, newTaskRun(wrID, uuid.New()))
	_ = r.Create(ctx, newTaskRun(uuid.New(), uuid.New()))

	list, err := r.ListByWorkflowRunID(ctx, wrID)
	if err != nil {
		t.Fatalf("ListByWorkflowRunID: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("ListByWorkflowRunID length: got %d, want 2", len(list))
	}
}

func TestTaskRunRepo_ListByTaskID(t *testing.T) {
	r := mock.NewTaskRunRepo()
	taskID := uuid.New()
	_ = r.Create(ctx, newTaskRun(uuid.New(), taskID))
	_ = r.Create(ctx, newTaskRun(uuid.New(), uuid.New()))

	list, err := r.ListByTaskID(ctx, taskID)
	if err != nil {
		t.Fatalf("ListByTaskID: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListByTaskID length: got %d, want 1", len(list))
	}
}

func TestTaskRunRepo_ListByStatus(t *testing.T) {
	r := mock.NewTaskRunRepo()
	_ = r.Create(ctx, newTaskRun(uuid.New(), uuid.New())) // pending
	tr2 := newTaskRun(uuid.New(), uuid.New())
	tr2.Status = domain.StatusRunning
	_ = r.Create(ctx, tr2)

	list, err := r.ListByStatus(ctx, domain.StatusRunning)
	if err != nil {
		t.Fatalf("ListByStatus: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListByStatus(running) length: got %d, want 1", len(list))
	}
}

// ── WorkerRepo ────────────────────────────────────────────────────────────────

func TestWorkerRepo_CreateAndGetByID(t *testing.T) {
	r := mock.NewWorkerRepo()
	w := newWorker()

	if err := r.Create(ctx, w); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := r.GetByID(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Hostname != w.Hostname {
		t.Errorf("Hostname: got %q, want %q", got.Hostname, w.Hostname)
	}
}

func TestWorkerRepo_GetByID_NotFound(t *testing.T) {
	r := mock.NewWorkerRepo()
	_, err := r.GetByID(ctx, uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestWorkerRepo_Update(t *testing.T) {
	r := mock.NewWorkerRepo()
	w := newWorker()
	_ = r.Create(ctx, w)

	w.Hostname = "host-updated"
	if err := r.Update(ctx, w); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := r.GetByID(ctx, w.ID)
	if got.Hostname != "host-updated" {
		t.Errorf("Hostname after Update: got %q, want %q", got.Hostname, "host-updated")
	}
}

func TestWorkerRepo_Delete(t *testing.T) {
	r := mock.NewWorkerRepo()
	w := newWorker()
	_ = r.Create(ctx, w)

	if err := r.Delete(ctx, w.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !errors.Is(r.Delete(ctx, w.ID), repository.ErrNotFound) {
		t.Error("expected ErrNotFound on second Delete")
	}
}

func TestWorkerRepo_ListActive(t *testing.T) {
	r := mock.NewWorkerRepo()
	active := newWorker()
	inactive := newWorker()
	inactive.Status = domain.WorkerStatusInactive
	_ = r.Create(ctx, active)
	_ = r.Create(ctx, inactive)

	list, err := r.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListActive length: got %d, want 1", len(list))
	}
	if list[0].ID != active.ID {
		t.Error("ListActive returned wrong worker")
	}
}

func TestWorkerRepo_UpdateHeartbeat(t *testing.T) {
	r := mock.NewWorkerRepo()
	w := newWorker()
	_ = r.Create(ctx, w)

	newBeat := time.Now().Add(time.Minute).UTC()
	if err := r.UpdateHeartbeat(ctx, w.ID, newBeat); err != nil {
		t.Fatalf("UpdateHeartbeat: %v", err)
	}
	got, _ := r.GetByID(ctx, w.ID)
	if !got.LastHeartbeat.Equal(newBeat) {
		t.Errorf("LastHeartbeat: got %v, want %v", got.LastHeartbeat, newBeat)
	}
}

func TestWorkerRepo_UpdateHeartbeat_NotFound(t *testing.T) {
	r := mock.NewWorkerRepo()
	err := r.UpdateHeartbeat(ctx, uuid.New(), time.Now())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ── interface compliance ──────────────────────────────────────────────────────

// These compile-time checks ensure each mock struct satisfies the corresponding
// repository interface.
var (
	_ repository.WorkflowRepository    = (*mock.WorkflowRepo)(nil)
	_ repository.TaskRepository        = (*mock.TaskRepo)(nil)
	_ repository.WorkflowRunRepository = (*mock.WorkflowRunRepo)(nil)
	_ repository.TaskRunRepository     = (*mock.TaskRunRepo)(nil)
	_ repository.WorkerRepository      = (*mock.WorkerRepo)(nil)
)
