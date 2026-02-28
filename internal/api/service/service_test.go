package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/api/service"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository/mock"
)

var ctx = context.Background()

// newService returns a Service backed by empty in-memory mock repositories.
func newService() *service.Service {
	return service.New(
		mock.NewWorkflowRepo(),
		mock.NewWorkflowRunRepo(),
		mock.NewTaskRunRepo(),
		mock.NewWorkerRepo(),
	)
}

// newServiceWithRepos returns a Service and its underlying mock repos for direct manipulation.
func newServiceWithRepos() (*service.Service, *mock.WorkflowRepo, *mock.WorkflowRunRepo, *mock.TaskRunRepo, *mock.WorkerRepo) {
	wfRepo := mock.NewWorkflowRepo()
	wrRepo := mock.NewWorkflowRunRepo()
	trRepo := mock.NewTaskRunRepo()
	wkRepo := mock.NewWorkerRepo()
	svc := service.New(wfRepo, wrRepo, trRepo, wkRepo)
	return svc, wfRepo, wrRepo, trRepo, wkRepo
}

// ── CreateWorkflow ────────────────────────────────────────────────────────────

func TestCreateWorkflow_Success(t *testing.T) {
	svc := newService()
	in := service.CreateWorkflowInput{
		Name:         "my-workflow",
		Description:  "desc",
		ScheduleCron: "0 * * * *",
		IsActive:     true,
	}
	wf, err := svc.CreateWorkflow(ctx, in)
	if err != nil {
		t.Fatalf("CreateWorkflow: %v", err)
	}
	if wf.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if wf.Name != in.Name {
		t.Errorf("Name: got %q, want %q", wf.Name, in.Name)
	}
	if wf.Description != in.Description {
		t.Errorf("Description: got %q, want %q", wf.Description, in.Description)
	}
	if wf.ScheduleCron != in.ScheduleCron {
		t.Errorf("ScheduleCron: got %q, want %q", wf.ScheduleCron, in.ScheduleCron)
	}
	if !wf.IsActive {
		t.Error("IsActive: expected true")
	}
	if wf.CreatedAt.IsZero() {
		t.Error("CreatedAt: expected non-zero")
	}
}

func TestCreateWorkflow_UniqueIDs(t *testing.T) {
	svc := newService()
	in := service.CreateWorkflowInput{Name: "wf"}
	wf1, _ := svc.CreateWorkflow(ctx, in)
	wf2, _ := svc.CreateWorkflow(ctx, in)
	if wf1.ID == wf2.ID {
		t.Error("expected unique IDs for each created workflow")
	}
}

// ── ListWorkflows ─────────────────────────────────────────────────────────────

func TestListWorkflows_Empty(t *testing.T) {
	svc := newService()
	list, err := svc.ListWorkflows(ctx, 0, 20)
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty slice, got %d items", len(list))
	}
}

func TestListWorkflows_All(t *testing.T) {
	svc, wfRepo, _, _, _ := newServiceWithRepos()
	for i := 0; i < 3; i++ {
		_ = wfRepo.Create(ctx, &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()})
	}
	list, err := svc.ListWorkflows(ctx, 0, 20)
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 items, got %d", len(list))
	}
}

func TestListWorkflows_Pagination(t *testing.T) {
	svc, wfRepo, _, _, _ := newServiceWithRepos()
	for i := 0; i < 5; i++ {
		_ = wfRepo.Create(ctx, &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()})
	}
	// offset=1, limit=2 → should return exactly 2 items
	list, err := svc.ListWorkflows(ctx, 1, 2)
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 items, got %d", len(list))
	}
}

func TestListWorkflows_OffsetBeyondEnd(t *testing.T) {
	svc, wfRepo, _, _, _ := newServiceWithRepos()
	_ = wfRepo.Create(ctx, &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()})

	list, err := svc.ListWorkflows(ctx, 10, 20)
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 items, got %d", len(list))
	}
}

func TestListWorkflows_NegativeOffset(t *testing.T) {
	svc, wfRepo, _, _, _ := newServiceWithRepos()
	for i := 0; i < 2; i++ {
		_ = wfRepo.Create(ctx, &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()})
	}
	list, err := svc.ListWorkflows(ctx, -1, 20)
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 items with negative offset, got %d", len(list))
	}
}

// ── TriggerWorkflow ───────────────────────────────────────────────────────────

func TestTriggerWorkflow_Success(t *testing.T) {
	svc, wfRepo, _, _, _ := newServiceWithRepos()
	wf := &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()}
	_ = wfRepo.Create(ctx, wf)

	run, err := svc.TriggerWorkflow(ctx, wf.ID)
	if err != nil {
		t.Fatalf("TriggerWorkflow: %v", err)
	}
	if run.ID == uuid.Nil {
		t.Error("expected non-nil run ID")
	}
	if run.WorkflowID != wf.ID {
		t.Errorf("WorkflowID: got %v, want %v", run.WorkflowID, wf.ID)
	}
	if run.Status != domain.StatusPending {
		t.Errorf("Status: got %q, want pending", run.Status)
	}
	if run.StartedAt.IsZero() {
		t.Error("StartedAt: expected non-zero")
	}
}

func TestTriggerWorkflow_NotFound(t *testing.T) {
	svc := newService()
	_, err := svc.TriggerWorkflow(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent workflow, got nil")
	}
	if !isErrNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// isErrNotFound checks whether err is the repository.ErrNotFound sentinel.
func isErrNotFound(err error) bool {
	return err == repository.ErrNotFound
}

// ── ListWorkflowRuns ──────────────────────────────────────────────────────────

func TestListWorkflowRuns_Empty(t *testing.T) {
	svc := newService()
	runs, err := svc.ListWorkflowRuns(ctx, "")
	if err != nil {
		t.Fatalf("ListWorkflowRuns: %v", err)
	}
	if runs == nil {
		runs = []*domain.WorkflowRun{}
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestListWorkflowRuns_ByStatus(t *testing.T) {
	svc, wfRepo, wrRepo, _, _ := newServiceWithRepos()
	wf := &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()}
	_ = wfRepo.Create(ctx, wf)

	pending := &domain.WorkflowRun{ID: uuid.New(), WorkflowID: wf.ID, Status: domain.StatusPending, StartedAt: time.Now().UTC()}
	running := &domain.WorkflowRun{ID: uuid.New(), WorkflowID: wf.ID, Status: domain.StatusRunning, StartedAt: time.Now().UTC()}
	_ = wrRepo.Create(ctx, pending)
	_ = wrRepo.Create(ctx, running)

	runs, err := svc.ListWorkflowRuns(ctx, domain.StatusPending)
	if err != nil {
		t.Fatalf("ListWorkflowRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 pending run, got %d", len(runs))
	}
	if runs[0].Status != domain.StatusPending {
		t.Errorf("Status: got %q, want pending", runs[0].Status)
	}
}

func TestListWorkflowRuns_All(t *testing.T) {
	svc, wfRepo, wrRepo, _, _ := newServiceWithRepos()
	wf := &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()}
	_ = wfRepo.Create(ctx, wf)
	_ = wrRepo.Create(ctx, &domain.WorkflowRun{ID: uuid.New(), WorkflowID: wf.ID, Status: domain.StatusPending, StartedAt: time.Now().UTC()})
	_ = wrRepo.Create(ctx, &domain.WorkflowRun{ID: uuid.New(), WorkflowID: wf.ID, Status: domain.StatusRunning, StartedAt: time.Now().UTC()})

	runs, err := svc.ListWorkflowRuns(ctx, "")
	if err != nil {
		t.Fatalf("ListWorkflowRuns: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

// ── ListTaskRuns ──────────────────────────────────────────────────────────────

func TestListTaskRuns_Empty(t *testing.T) {
	svc := newService()
	trs, err := svc.ListTaskRuns(ctx, "")
	if err != nil {
		t.Fatalf("ListTaskRuns: %v", err)
	}
	if trs == nil {
		trs = []*domain.TaskRun{}
	}
	if len(trs) != 0 {
		t.Errorf("expected 0 task runs, got %d", len(trs))
	}
}

func TestListTaskRuns_ByStatus(t *testing.T) {
	svc, wfRepo, wrRepo, trRepo, _ := newServiceWithRepos()
	wf := &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()}
	_ = wfRepo.Create(ctx, wf)
	wr := &domain.WorkflowRun{ID: uuid.New(), WorkflowID: wf.ID, Status: domain.StatusRunning, StartedAt: time.Now().UTC()}
	_ = wrRepo.Create(ctx, wr)

	taskID := uuid.New()
	_ = trRepo.Create(ctx, &domain.TaskRun{ID: uuid.New(), WorkflowRunID: wr.ID, TaskID: taskID, Status: domain.StatusPending, Attempt: 1, StartedAt: time.Now().UTC()})
	_ = trRepo.Create(ctx, &domain.TaskRun{ID: uuid.New(), WorkflowRunID: wr.ID, TaskID: taskID, Status: domain.StatusRunning, Attempt: 1, StartedAt: time.Now().UTC()})

	trs, err := svc.ListTaskRuns(ctx, domain.StatusRunning)
	if err != nil {
		t.Fatalf("ListTaskRuns: %v", err)
	}
	if len(trs) != 1 {
		t.Errorf("expected 1 running task run, got %d", len(trs))
	}
}

func TestListTaskRuns_All(t *testing.T) {
	svc, wfRepo, wrRepo, trRepo, _ := newServiceWithRepos()
	wf := &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()}
	_ = wfRepo.Create(ctx, wf)
	wr := &domain.WorkflowRun{ID: uuid.New(), WorkflowID: wf.ID, Status: domain.StatusRunning, StartedAt: time.Now().UTC()}
	_ = wrRepo.Create(ctx, wr)

	taskID := uuid.New()
	_ = trRepo.Create(ctx, &domain.TaskRun{ID: uuid.New(), WorkflowRunID: wr.ID, TaskID: taskID, Status: domain.StatusPending, Attempt: 1, StartedAt: time.Now().UTC()})
	_ = trRepo.Create(ctx, &domain.TaskRun{ID: uuid.New(), WorkflowRunID: wr.ID, TaskID: taskID, Status: domain.StatusSuccess, Attempt: 1, StartedAt: time.Now().UTC()})

	trs, err := svc.ListTaskRuns(ctx, "")
	if err != nil {
		t.Fatalf("ListTaskRuns: %v", err)
	}
	if len(trs) != 2 {
		t.Errorf("expected 2 task runs, got %d", len(trs))
	}
}

// ── ListWorkers ───────────────────────────────────────────────────────────────

func TestListWorkers_Empty(t *testing.T) {
	svc := newService()
	workers, err := svc.ListWorkers(ctx)
	if err != nil {
		t.Fatalf("ListWorkers: %v", err)
	}
	if len(workers) != 0 {
		t.Errorf("expected 0 workers, got %d", len(workers))
	}
}

func TestListWorkers_OnlyActive(t *testing.T) {
	svc, _, _, _, wkRepo := newServiceWithRepos()
	active := &domain.Worker{ID: uuid.New(), Hostname: "host-1", LastHeartbeat: time.Now().UTC(), Status: domain.WorkerStatusActive}
	inactive := &domain.Worker{ID: uuid.New(), Hostname: "host-2", LastHeartbeat: time.Now().UTC(), Status: domain.WorkerStatusInactive}
	_ = wkRepo.Create(ctx, active)
	_ = wkRepo.Create(ctx, inactive)

	workers, err := svc.ListWorkers(ctx)
	if err != nil {
		t.Fatalf("ListWorkers: %v", err)
	}
	if len(workers) != 1 {
		t.Errorf("expected 1 active worker, got %d", len(workers))
	}
	if workers[0].ID != active.ID {
		t.Errorf("unexpected worker returned: got %v, want %v", workers[0].ID, active.ID)
	}
}
