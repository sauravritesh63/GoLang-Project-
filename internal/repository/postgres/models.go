// Package postgres provides GORM-based implementations of the repository
// interfaces defined in the parent package. All repositories accept a
// *gorm.DB via their constructors (dependency injection — no package-level
// globals) and propagate the caller's context to every DB call.
package postgres

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
)

// ── Workflow ──────────────────────────────────────────────────────────────────

type workflowModel struct {
	ID           string    `gorm:"type:uuid;primaryKey;column:id"`
	Name         string    `gorm:"column:name;not null"`
	Description  string    `gorm:"column:description;not null;default:''"`
	ScheduleCron string    `gorm:"column:schedule_cron;not null;default:''"`
	IsActive     bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (workflowModel) TableName() string { return "workflows" }

func (m *workflowModel) toDomain() (*domain.Workflow, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, fmt.Errorf("workflow: invalid id %q: %w", m.ID, err)
	}
	return &domain.Workflow{
		ID:           id,
		Name:         m.Name,
		Description:  m.Description,
		ScheduleCron: m.ScheduleCron,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
	}, nil
}

func workflowFromDomain(wf *domain.Workflow) *workflowModel {
	return &workflowModel{
		ID:           wf.ID.String(),
		Name:         wf.Name,
		Description:  wf.Description,
		ScheduleCron: wf.ScheduleCron,
		IsActive:     wf.IsActive,
		CreatedAt:    wf.CreatedAt,
	}
}

// ── Task ──────────────────────────────────────────────────────────────────────

type taskModel struct {
	ID                string    `gorm:"type:uuid;primaryKey;column:id"`
	WorkflowID        string    `gorm:"type:uuid;column:workflow_id;not null"`
	Name              string    `gorm:"column:name;not null"`
	Command           string    `gorm:"column:command;not null;default:''"`
	RetryCount        int       `gorm:"column:retry_count;not null;default:0"`
	RetryDelaySeconds int       `gorm:"column:retry_delay_seconds;not null;default:0"`
	TimeoutSeconds    int       `gorm:"column:timeout_seconds;not null;default:0"`
	CreatedAt         time.Time `gorm:"column:created_at;not null"`
}

func (taskModel) TableName() string { return "tasks" }

func (m *taskModel) toDomain() (*domain.Task, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, fmt.Errorf("task: invalid id %q: %w", m.ID, err)
	}
	wfID, err := uuid.Parse(m.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("task: invalid workflow_id %q: %w", m.WorkflowID, err)
	}
	return &domain.Task{
		ID:                id,
		WorkflowID:        wfID,
		Name:              m.Name,
		Command:           m.Command,
		RetryCount:        m.RetryCount,
		RetryDelaySeconds: m.RetryDelaySeconds,
		TimeoutSeconds:    m.TimeoutSeconds,
		CreatedAt:         m.CreatedAt,
	}, nil
}

func taskFromDomain(t *domain.Task) *taskModel {
	return &taskModel{
		ID:                t.ID.String(),
		WorkflowID:        t.WorkflowID.String(),
		Name:              t.Name,
		Command:           t.Command,
		RetryCount:        t.RetryCount,
		RetryDelaySeconds: t.RetryDelaySeconds,
		TimeoutSeconds:    t.TimeoutSeconds,
		CreatedAt:         t.CreatedAt,
	}
}

// ── WorkflowRun ───────────────────────────────────────────────────────────────

type workflowRunModel struct {
	ID         string     `gorm:"type:uuid;primaryKey;column:id"`
	WorkflowID string     `gorm:"type:uuid;column:workflow_id;not null"`
	Status     string     `gorm:"column:status;not null;default:'pending'"`
	StartedAt  time.Time  `gorm:"column:started_at;not null"`
	FinishedAt *time.Time `gorm:"column:finished_at"`
}

func (workflowRunModel) TableName() string { return "workflow_runs" }

func (m *workflowRunModel) toDomain() (*domain.WorkflowRun, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, fmt.Errorf("workflow_run: invalid id %q: %w", m.ID, err)
	}
	wfID, err := uuid.Parse(m.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow_run: invalid workflow_id %q: %w", m.WorkflowID, err)
	}
	return &domain.WorkflowRun{
		ID:         id,
		WorkflowID: wfID,
		Status:     domain.Status(m.Status),
		StartedAt:  m.StartedAt,
		FinishedAt: m.FinishedAt,
	}, nil
}

func workflowRunFromDomain(wr *domain.WorkflowRun) *workflowRunModel {
	return &workflowRunModel{
		ID:         wr.ID.String(),
		WorkflowID: wr.WorkflowID.String(),
		Status:     string(wr.Status),
		StartedAt:  wr.StartedAt,
		FinishedAt: wr.FinishedAt,
	}
}

// ── TaskRun ───────────────────────────────────────────────────────────────────

type taskRunModel struct {
	ID            string     `gorm:"type:uuid;primaryKey;column:id"`
	WorkflowRunID string     `gorm:"type:uuid;column:workflow_run_id;not null"`
	TaskID        string     `gorm:"type:uuid;column:task_id;not null"`
	Status        string     `gorm:"column:status;not null;default:'pending'"`
	Attempt       int        `gorm:"column:attempt;not null;default:1"`
	StartedAt     time.Time  `gorm:"column:started_at;not null"`
	FinishedAt    *time.Time `gorm:"column:finished_at"`
	Logs          string     `gorm:"column:logs;not null;default:''"`
}

func (taskRunModel) TableName() string { return "task_runs" }

func (m *taskRunModel) toDomain() (*domain.TaskRun, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, fmt.Errorf("task_run: invalid id %q: %w", m.ID, err)
	}
	wrID, err := uuid.Parse(m.WorkflowRunID)
	if err != nil {
		return nil, fmt.Errorf("task_run: invalid workflow_run_id %q: %w", m.WorkflowRunID, err)
	}
	tID, err := uuid.Parse(m.TaskID)
	if err != nil {
		return nil, fmt.Errorf("task_run: invalid task_id %q: %w", m.TaskID, err)
	}
	return &domain.TaskRun{
		ID:            id,
		WorkflowRunID: wrID,
		TaskID:        tID,
		Status:        domain.Status(m.Status),
		Attempt:       m.Attempt,
		StartedAt:     m.StartedAt,
		FinishedAt:    m.FinishedAt,
		Logs:          m.Logs,
	}, nil
}

func taskRunFromDomain(tr *domain.TaskRun) *taskRunModel {
	return &taskRunModel{
		ID:            tr.ID.String(),
		WorkflowRunID: tr.WorkflowRunID.String(),
		TaskID:        tr.TaskID.String(),
		Status:        string(tr.Status),
		Attempt:       tr.Attempt,
		StartedAt:     tr.StartedAt,
		FinishedAt:    tr.FinishedAt,
		Logs:          tr.Logs,
	}
}

// ── Worker ────────────────────────────────────────────────────────────────────

type workerModel struct {
	ID            string    `gorm:"type:uuid;primaryKey;column:id"`
	Hostname      string    `gorm:"column:hostname;not null"`
	LastHeartbeat time.Time `gorm:"column:last_heartbeat;not null"`
	Status        string    `gorm:"column:status;not null;default:'active'"`
}

func (workerModel) TableName() string { return "workers" }

func (m *workerModel) toDomain() (*domain.Worker, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, fmt.Errorf("worker: invalid id %q: %w", m.ID, err)
	}
	return &domain.Worker{
		ID:            id,
		Hostname:      m.Hostname,
		LastHeartbeat: m.LastHeartbeat,
		Status:        domain.WorkerStatus(m.Status),
	}, nil
}

func workerFromDomain(w *domain.Worker) *workerModel {
	return &workerModel{
		ID:            w.ID.String(),
		Hostname:      w.Hostname,
		LastHeartbeat: w.LastHeartbeat,
		Status:        string(w.Status),
	}
}

