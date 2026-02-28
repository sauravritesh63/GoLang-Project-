package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/api/handler"
	"github.com/sauravritesh63/GoLang-Project-/internal/api/service"
	ws "github.com/sauravritesh63/GoLang-Project-/internal/api/websocket"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestRouter builds a fully wired Gin engine backed by in-memory mock repos.
func newTestRouter() (*gin.Engine, *mock.WorkflowRepo, *mock.WorkflowRunRepo, *mock.TaskRunRepo, *mock.WorkerRepo) {
	wfRepo := mock.NewWorkflowRepo()
	wrRepo := mock.NewWorkflowRunRepo()
	trRepo := mock.NewTaskRunRepo()
	wkRepo := mock.NewWorkerRepo()

	svc := service.New(wfRepo, wrRepo, trRepo, wkRepo)
	hub := ws.NewHub()
	h := handler.New(svc, hub)

	r := gin.New()
	h.RegisterRoutes(r)
	return r, wfRepo, wrRepo, trRepo, wkRepo
}

// TestCreateWorkflow_Success verifies POST /workflows returns 201 with the
// newly created workflow.
func TestCreateWorkflow_Success(t *testing.T) {
	r, _, _, _, _ := newTestRouter()

	body := `{"name":"my-workflow","description":"desc","schedule_cron":"0 * * * *","is_active":true}`
	req := httptest.NewRequest(http.MethodPost, "/workflows", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var wf domain.Workflow
	if err := json.NewDecoder(w.Body).Decode(&wf); err != nil {
		t.Fatal(err)
	}
	if wf.Name != "my-workflow" {
		t.Errorf("expected name 'my-workflow', got %q", wf.Name)
	}
	if wf.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

// TestCreateWorkflow_MissingName verifies POST /workflows returns 400 when
// the required 'name' field is absent.
func TestCreateWorkflow_MissingName(t *testing.T) {
	r, _, _, _, _ := newTestRouter()

	body := `{"description":"no name"}`
	req := httptest.NewRequest(http.MethodPost, "/workflows", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestListWorkflows_Empty verifies GET /workflows returns an empty JSON array
// when no workflows exist.
func TestListWorkflows_Empty(t *testing.T) {
	r, _, _, _, _ := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/workflows", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []domain.Workflow
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

// TestListWorkflows_Pagination verifies GET /workflows respects offset and limit.
func TestListWorkflows_Pagination(t *testing.T) {
	r, wfRepo, _, _, _ := newTestRouter()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_ = wfRepo.Create(ctx, &domain.Workflow{
			ID:        uuid.New(),
			Name:      "wf",
			CreatedAt: time.Now().UTC(),
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/workflows?offset=1&limit=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []domain.Workflow
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

// TestTriggerWorkflow_Success verifies POST /workflows/{id}/trigger returns 201.
func TestTriggerWorkflow_Success(t *testing.T) {
	r, wfRepo, _, _, _ := newTestRouter()
	ctx := context.Background()

	wf := &domain.Workflow{ID: uuid.New(), Name: "wf", CreatedAt: time.Now().UTC()}
	_ = wfRepo.Create(ctx, wf)

	req := httptest.NewRequest(http.MethodPost, "/workflows/"+wf.ID.String()+"/trigger", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var run domain.WorkflowRun
	if err := json.NewDecoder(w.Body).Decode(&run); err != nil {
		t.Fatal(err)
	}
	if run.WorkflowID != wf.ID {
		t.Errorf("expected workflow_id %s, got %s", wf.ID, run.WorkflowID)
	}
	if run.Status != domain.StatusPending {
		t.Errorf("expected status 'pending', got %q", run.Status)
	}
}

// TestTriggerWorkflow_NotFound verifies that triggering a non-existent workflow
// returns 404.
func TestTriggerWorkflow_NotFound(t *testing.T) {
	r, _, _, _, _ := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/workflows/"+uuid.New().String()+"/trigger", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestTriggerWorkflow_InvalidID verifies that a malformed UUID returns 400.
func TestTriggerWorkflow_InvalidID(t *testing.T) {
	r, _, _, _, _ := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/workflows/not-a-uuid/trigger", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestListWorkflowRuns_Empty verifies GET /workflow-runs returns an empty JSON
// array when no runs exist.
func TestListWorkflowRuns_Empty(t *testing.T) {
	r, _, _, _, _ := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/workflow-runs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// TestListTaskRuns_Empty verifies GET /task-runs returns an empty JSON array.
func TestListTaskRuns_Empty(t *testing.T) {
	r, _, _, _, _ := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/task-runs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// TestListWorkers_Active verifies GET /workers returns only active workers.
func TestListWorkers_Active(t *testing.T) {
	r, _, _, _, wkRepo := newTestRouter()
	ctx := context.Background()

	active := &domain.Worker{
		ID:            uuid.New(),
		Hostname:      "host-1",
		LastHeartbeat: time.Now().UTC(),
		Status:        domain.WorkerStatusActive,
	}
	inactive := &domain.Worker{
		ID:            uuid.New(),
		Hostname:      "host-2",
		LastHeartbeat: time.Now().UTC(),
		Status:        domain.WorkerStatusInactive,
	}
	_ = wkRepo.Create(ctx, active)
	_ = wkRepo.Create(ctx, inactive)

	req := httptest.NewRequest(http.MethodGet, "/workers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var workers []domain.Worker
	if err := json.NewDecoder(w.Body).Decode(&workers); err != nil {
		t.Fatal(err)
	}
	if len(workers) != 1 {
		t.Errorf("expected 1 active worker, got %d", len(workers))
	}
	if workers[0].Hostname != "host-1" {
		t.Errorf("expected host-1, got %q", workers[0].Hostname)
	}
}
