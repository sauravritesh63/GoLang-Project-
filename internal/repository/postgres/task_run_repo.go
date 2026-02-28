package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
	"gorm.io/gorm"
)

// TaskRunRepo is a GORM-backed implementation of repository.TaskRunRepository.
type TaskRunRepo struct {
	db *gorm.DB
}

// NewTaskRunRepo constructs a TaskRunRepo with the supplied *gorm.DB.
func NewTaskRunRepo(db *gorm.DB) *TaskRunRepo {
	return &TaskRunRepo{db: db}
}

func (r *TaskRunRepo) Create(ctx context.Context, tr *domain.TaskRun) error {
	return r.db.WithContext(ctx).Create(taskRunFromDomain(tr)).Error
}

func (r *TaskRunRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.TaskRun, error) {
	var m taskRunModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.String()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain()
}

func (r *TaskRunRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, finishedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":      string(status),
		"finished_at": finishedAt,
	}
	result := r.db.WithContext(ctx).
		Model(&taskRunModel{}).
		Where("id = ?", id.String()).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *TaskRunRepo) ListByWorkflowRunID(ctx context.Context, workflowRunID uuid.UUID) ([]*domain.TaskRun, error) {
	var models []taskRunModel
	if err := r.db.WithContext(ctx).
		Where("workflow_run_id = ?", workflowRunID.String()).
		Order("started_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.TaskRun, len(models))
	for i := range models {
		tr, err := models[i].toDomain()
		if err != nil {
			return nil, err
		}
		out[i] = tr
	}
	return out, nil
}

func (r *TaskRunRepo) ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskRun, error) {
	var models []taskRunModel
	if err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID.String()).
		Order("started_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.TaskRun, len(models))
	for i := range models {
		tr, err := models[i].toDomain()
		if err != nil {
			return nil, err
		}
		out[i] = tr
	}
	return out, nil
}

func (r *TaskRunRepo) ListByStatus(ctx context.Context, status domain.Status) ([]*domain.TaskRun, error) {
	var models []taskRunModel
	if err := r.db.WithContext(ctx).
		Where("status = ?", string(status)).
		Order("started_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.TaskRun, len(models))
	for i := range models {
		tr, err := models[i].toDomain()
		if err != nil {
			return nil, err
		}
		out[i] = tr
	}
	return out, nil
}
