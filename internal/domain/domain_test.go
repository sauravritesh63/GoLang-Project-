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

func TestWorkflowInstantiation(t *testing.T) {
id := uuid.New()
now := time.Now().UTC()
wf := domain.Workflow{
ID:           id,
Name:         "etl-pipeline",
Description:  "Daily ETL",
ScheduleCron: "0 2 * * *",
IsActive:     true,
CreatedAt:    now,
}

if wf.ID != id {
t.Errorf("ID mismatch: got %v, want %v", wf.ID, id)
}
if wf.Name != "etl-pipeline" {
t.Errorf("Name mismatch: got %q, want %q", wf.Name, "etl-pipeline")
}
if !wf.IsActive {
t.Error("IsActive should be true")
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

func TestTaskInstantiation(t *testing.T) {
task := domain.Task{
ID:                uuid.New(),
WorkflowID:        uuid.New(),
Name:              "extract",
Command:           "python extract.py",
RetryCount:        3,
RetryDelaySeconds: 5,
TimeoutSeconds:    60,
CreatedAt:         time.Now().UTC(),
}

if task.Name != "extract" {
t.Errorf("Name mismatch: got %q, want %q", task.Name, "extract")
}
if task.RetryCount != 3 {
t.Errorf("RetryCount mismatch: got %d, want %d", task.RetryCount, 3)
}
if task.TimeoutSeconds != 60 {
t.Errorf("TimeoutSeconds mismatch: got %d, want %d", task.TimeoutSeconds, 60)
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

func TestTaskDependencyInstantiation(t *testing.T) {
taskID := uuid.New()
depID := uuid.New()
td := domain.TaskDependency{
ID:              uuid.New(),
TaskID:          taskID,
DependsOnTaskID: depID,
}

if td.TaskID != taskID {
t.Errorf("TaskID mismatch: got %v, want %v", td.TaskID, taskID)
}
if td.DependsOnTaskID != depID {
t.Errorf("DependsOnTaskID mismatch: got %v, want %v", td.DependsOnTaskID, depID)
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

func TestWorkflowRunInstantiation(t *testing.T) {
wr := domain.WorkflowRun{
ID:         uuid.New(),
WorkflowID: uuid.New(),
Status:     domain.StatusRunning,
StartedAt:  time.Now().UTC(),
}

if wr.Status != domain.StatusRunning {
t.Errorf("Status mismatch: got %q, want %q", wr.Status, domain.StatusRunning)
}
if wr.FinishedAt != nil {
t.Error("FinishedAt should be nil for a running workflow")
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

var m map[string]interface{}
if err := json.Unmarshal(b, &m); err != nil {
t.Fatalf("unmarshal error: %v", err)
}
if _, ok := m["finished_at"]; ok {
t.Error("finished_at should be omitted when nil")
}
}

func TestWorkflowRunWithFinishedAt(t *testing.T) {
now := time.Now().UTC().Truncate(time.Second)
finished := now.Add(10 * time.Minute)
wr := domain.WorkflowRun{
ID:         uuid.New(),
WorkflowID: uuid.New(),
Status:     domain.StatusSuccess,
StartedAt:  now,
FinishedAt: &finished,
}

b, err := json.Marshal(wr)
if err != nil {
t.Fatalf("marshal error: %v", err)
}

var got domain.WorkflowRun
if err := json.Unmarshal(b, &got); err != nil {
t.Fatalf("unmarshal error: %v", err)
}

if got.FinishedAt == nil {
t.Fatal("FinishedAt should not be nil after round-trip")
}
if !got.FinishedAt.Equal(finished) {
t.Errorf("FinishedAt mismatch: got %v, want %v", got.FinishedAt, finished)
}
}

func TestTaskRunInstantiation(t *testing.T) {
tr := domain.TaskRun{
ID:            uuid.New(),
WorkflowRunID: uuid.New(),
TaskID:        uuid.New(),
Status:        domain.StatusPending,
Attempt:       1,
StartedAt:     time.Now().UTC(),
Logs:          "",
}

if tr.Status != domain.StatusPending {
t.Errorf("Status mismatch: got %q, want %q", tr.Status, domain.StatusPending)
}
if tr.Attempt != 1 {
t.Errorf("Attempt mismatch: got %d, want %d", tr.Attempt, 1)
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

func TestWorkerInstantiation(t *testing.T) {
w := domain.Worker{
ID:            uuid.New(),
Hostname:      "worker-1.local",
LastHeartbeat: time.Now().UTC(),
Status:        domain.WorkerStatusActive,
}

if w.Hostname != "worker-1.local" {
t.Errorf("Hostname mismatch: got %q, want %q", w.Hostname, "worker-1.local")
}
if w.Status != domain.WorkerStatusActive {
t.Errorf("Status mismatch: got %q, want %q", w.Status, domain.WorkerStatusActive)
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

func TestStatusJSONField(t *testing.T) {
wr := domain.WorkflowRun{
ID:         uuid.New(),
WorkflowID: uuid.New(),
Status:     domain.StatusFailed,
StartedAt:  time.Now().UTC(),
}

b, err := json.Marshal(wr)
if err != nil {
t.Fatalf("marshal error: %v", err)
}

var m map[string]interface{}
if err := json.Unmarshal(b, &m); err != nil {
t.Fatalf("unmarshal error: %v", err)
}

statusVal, ok := m["status"]
if !ok {
t.Fatal("status field missing from JSON")
}
if statusVal != "failed" {
t.Errorf("status JSON value: got %q, want %q", statusVal, "failed")
}
}
