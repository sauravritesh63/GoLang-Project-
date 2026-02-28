// Package service provides the API business-logic layer for the distributed
// task scheduler. It mediates between HTTP handlers and the repository layer.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
)

// Service holds all repository dependencies and exposes use-case methods
// consumed by the HTTP handler layer.
type Service struct {
	workflows    repository.WorkflowRepository
	workflowRuns repository.WorkflowRunRepository
	taskRuns     repository.TaskRunRepository
	workers      repository.WorkerRepository
}

// New creates a Service with the supplied repository implementations.
func New(
	workflows repository.WorkflowRepository,
	workflowRuns repository.WorkflowRunRepository,
	taskRuns repository.TaskRunRepository,
	workers repository.WorkerRepository,
) *Service {
	return &Service{
		workflows:    workflows,
		workflowRuns: workflowRuns,
		taskRuns:     taskRuns,
		workers:      workers,
	}
}

// CreateWorkflowInput carries the fields supplied by the caller when creating
// a new workflow. ID and CreatedAt are generated here.
type CreateWorkflowInput struct {
	Name         string `json:"name"         binding:"required"`
	Description  string `json:"description"`
	ScheduleCron string `json:"schedule_cron"`
	IsActive     bool   `json:"is_active"`
}

// CreateWorkflow persists a new workflow and returns the stored entity.
func (s *Service) CreateWorkflow(ctx context.Context, in CreateWorkflowInput) (*domain.Workflow, error) {
	wf := &domain.Workflow{
		ID:           uuid.New(),
		Name:         in.Name,
		Description:  in.Description,
		ScheduleCron: in.ScheduleCron,
		IsActive:     in.IsActive,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.workflows.Create(ctx, wf); err != nil {
		return nil, err
	}
	return wf, nil
}

// ListWorkflows returns all workflows. Pagination (offset/limit) is applied
// in-process because the repository List method returns all records.
func (s *Service) ListWorkflows(ctx context.Context, offset, limit int) ([]*domain.Workflow, error) {
	all, err := s.workflows.List(ctx)
	if err != nil {
		return nil, err
	}
	return paginate(all, offset, limit), nil
}

// TriggerWorkflow creates a new WorkflowRun for the given workflow ID.
func (s *Service) TriggerWorkflow(ctx context.Context, workflowID uuid.UUID) (*domain.WorkflowRun, error) {
	// Verify the workflow exists.
	if _, err := s.workflows.GetByID(ctx, workflowID); err != nil {
		return nil, err
	}
	run := &domain.WorkflowRun{
		ID:         uuid.New(),
		WorkflowID: workflowID,
		Status:     domain.StatusPending,
		StartedAt:  time.Now().UTC(),
	}
	if err := s.workflowRuns.Create(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

// ListWorkflowRuns returns all workflow runs, optionally filtered by status.
func (s *Service) ListWorkflowRuns(ctx context.Context, status domain.Status) ([]*domain.WorkflowRun, error) {
	if status != "" {
		return s.workflowRuns.ListByStatus(ctx, status)
	}
	// No status filter — collect runs for all workflows.
	wfs, err := s.workflows.List(ctx)
	if err != nil {
		return nil, err
	}
	var runs []*domain.WorkflowRun
	for _, wf := range wfs {
		r, err := s.workflowRuns.ListByWorkflowID(ctx, wf.ID)
		if err != nil {
			return nil, err
		}
		runs = append(runs, r...)
	}
	return runs, nil
}

// ListTaskRuns returns all task runs, optionally filtered by status.
func (s *Service) ListTaskRuns(ctx context.Context, status domain.Status) ([]*domain.TaskRun, error) {
	if status != "" {
		return s.taskRuns.ListByStatus(ctx, status)
	}
	// No status filter — collect task runs for all workflow runs across all workflows.
	wfs, err := s.workflows.List(ctx)
	if err != nil {
		return nil, err
	}
	var taskRuns []*domain.TaskRun
	for _, wf := range wfs {
		wfRuns, err := s.workflowRuns.ListByWorkflowID(ctx, wf.ID)
		if err != nil {
			return nil, err
		}
		for _, wr := range wfRuns {
			trs, err := s.taskRuns.ListByWorkflowRunID(ctx, wr.ID)
			if err != nil {
				return nil, err
			}
			taskRuns = append(taskRuns, trs...)
		}
	}
	return taskRuns, nil
}

// ListWorkers returns all active workers.
func (s *Service) ListWorkers(ctx context.Context) ([]*domain.Worker, error) {
	return s.workers.ListActive(ctx)
}

// paginate applies offset/limit slicing to a slice; non-positive limit means
// return all remaining items. A negative offset is treated as zero.
func paginate[T any](items []T, offset, limit int) []T {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(items) {
		return []T{}
	}
	items = items[offset:]
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items
}
