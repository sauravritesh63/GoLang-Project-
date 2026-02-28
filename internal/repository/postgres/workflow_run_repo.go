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

// WorkflowRunRepo is a GORM-backed implementation of repository.WorkflowRunRepository.
type WorkflowRunRepo struct {
	db *gorm.DB
}

// NewWorkflowRunRepo constructs a WorkflowRunRepo with the supplied *gorm.DB.
func NewWorkflowRunRepo(db *gorm.DB) *WorkflowRunRepo {
	return &WorkflowRunRepo{db: db}
}

func (r *WorkflowRunRepo) Create(ctx context.Context, wr *domain.WorkflowRun) error {
	return r.db.WithContext(ctx).Create(workflowRunFromDomain(wr)).Error
}

func (r *WorkflowRunRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowRun, error) {
	var m workflowRunModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.String()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain()
}

func (r *WorkflowRunRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, finishedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":      string(status),
		"finished_at": finishedAt,
	}
	result := r.db.WithContext(ctx).
		Model(&workflowRunModel{}).
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

func (r *WorkflowRunRepo) ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.WorkflowRun, error) {
	var models []workflowRunModel
	if err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID.String()).
		Order("started_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.WorkflowRun, len(models))
	for i := range models {
		wr, err := models[i].toDomain()
		if err != nil {
			return nil, err
		}
		out[i] = wr
	}
	return out, nil
}

func (r *WorkflowRunRepo) ListByStatus(ctx context.Context, status domain.Status) ([]*domain.WorkflowRun, error) {
	var models []workflowRunModel
	if err := r.db.WithContext(ctx).
		Where("status = ?", string(status)).
		Order("started_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.WorkflowRun, len(models))
	for i := range models {
		wr, err := models[i].toDomain()
		if err != nil {
			return nil, err
		}
		out[i] = wr
	}
	return out, nil
}
