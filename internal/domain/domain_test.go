package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
)

func TestStatusConstants(t *testing.T) {
	cases := []struct {
		name string
		s    domain.Status
		want string
	}{
		{"pending", domain.StatusPending, "pending"},
		{"running", domain.StatusRunning, "running"},
		{"success", domain.StatusSuccess, "success"},
		{"failed", domain.StatusFailed, "failed"},
	}
	for _, tc := range cases {
		if string(tc.s) != tc.want {
			t.Errorf("Status %s: got %q, want %q", tc.name, tc.s, tc.want)
		}
	}
}

func TestWorkerStatusConstants(t *testing.T) {
	if string(domain.WorkerStatusActive) != "active" {
		t.Errorf("WorkerStatusActive: got %q, want %q", domain.WorkerStatusActive, "active")
	}
	if string(domain.WorkerStatusInactive) != "inactive" {
		t.Errorf("WorkerStatusInactive: got %q, want %q", domain.WorkerStatusInactive, "inactive")
	}
}

func TestWorkflowJSONRoundtrip(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	wf := domain.Workflow{
		ID:           id,
		Name:         "etl-pipeline",
		Description:  "Daily ETL",
		ScheduleCron: "0 2 * * *",
		IsActive:     true,
		CreatedAt:    now,
	}

	b, err := json.Marshal(wf)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got domain.Workflow
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.ID != wf.ID {
		t.Errorf("ID mismatch: got %v, want %v", got.ID, wf.ID)
	}
	if got.Name != wf.Name {
		t.Errorf("Name mismatch: got %q, want %q", got.Name, wf.Name)
	}
	if got.ScheduleCron != wf.ScheduleCron {
		t.Errorf("ScheduleCron mismatch: got %q, want %q", got.ScheduleCron, wf.ScheduleCron)
	}
	if got.IsActive != wf.IsActive {
		t.Errorf("IsActive mismatch: got %v, want %v", got.IsActive, wf.IsActive)
	}
}

func TestTaskJSONRoundtrip(t *testing.T) {
	task := domain.Task{
		ID:                uuid.New(),
		WorkflowID:        uuid.New(),
		Name:              "extract",
		Command:           "python extract.py",
		RetryCount:        3,
		RetryDelaySeconds: 5,
		TimeoutSeconds:    60,
		CreatedAt:         time.Now().UTC().Truncate(time.Second),
	}

	b, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got domain.Task
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.RetryCount != task.RetryCount {
		t.Errorf("RetryCount mismatch: got %d, want %d", got.RetryCount, task.RetryCount)
	}
	if got.Command != task.Command {
		t.Errorf("Command mismatch: got %q, want %q", got.Command, task.Command)
	}
}

func TestWorkflowRunOptionalFinishedAt(t *testing.T) {
	wr := domain.WorkflowRun{
		ID:         uuid.New(),
		WorkflowID: uuid.New(),
		Status:     domain.StatusRunning,
		StartedAt:  time.Now().UTC(),
	}

	b, err := json.Marshal(wr)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// finished_at should be omitted when nil
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if _, ok := m["finished_at"]; ok {
		t.Error("finished_at should be omitted when nil")
	}
}

func TestTaskRunOptionalFinishedAt(t *testing.T) {
	tr := domain.TaskRun{
		ID:            uuid.New(),
		WorkflowRunID: uuid.New(),
		TaskID:        uuid.New(),
		Status:        domain.StatusPending,
		Attempt:       1,
		StartedAt:     time.Now().UTC(),
		Logs:          "",
	}

	b, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if _, ok := m["finished_at"]; ok {
		t.Error("finished_at should be omitted when nil")
	}
}

func TestTaskDependencyJSONRoundtrip(t *testing.T) {
	td := domain.TaskDependency{
		ID:              uuid.New(),
		TaskID:          uuid.New(),
		DependsOnTaskID: uuid.New(),
	}

	b, err := json.Marshal(td)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got domain.TaskDependency
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.TaskID != td.TaskID {
		t.Errorf("TaskID mismatch: got %v, want %v", got.TaskID, td.TaskID)
	}
	if got.DependsOnTaskID != td.DependsOnTaskID {
		t.Errorf("DependsOnTaskID mismatch: got %v, want %v", got.DependsOnTaskID, td.DependsOnTaskID)
	}
}

func TestWorkerJSONRoundtrip(t *testing.T) {
	w := domain.Worker{
		ID:            uuid.New(),
		Hostname:      "worker-1.local",
		LastHeartbeat: time.Now().UTC().Truncate(time.Second),
		Status:        domain.WorkerStatusActive,
	}

	b, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got domain.Worker
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.Hostname != w.Hostname {
		t.Errorf("Hostname mismatch: got %q, want %q", got.Hostname, w.Hostname)
	}
	if got.Status != w.Status {
		t.Errorf("Status mismatch: got %q, want %q", got.Status, w.Status)
	}
}
