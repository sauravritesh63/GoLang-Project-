package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
	"gorm.io/gorm"
)

// WorkflowRepo is a GORM-backed implementation of repository.WorkflowRepository.
type WorkflowRepo struct {
	db *gorm.DB
}

// NewWorkflowRepo constructs a WorkflowRepo with the supplied *gorm.DB.
func NewWorkflowRepo(db *gorm.DB) *WorkflowRepo {
	return &WorkflowRepo{db: db}
}

func (r *WorkflowRepo) Create(ctx context.Context, wf *domain.Workflow) error {
	return r.db.WithContext(ctx).Create(workflowFromDomain(wf)).Error
}

func (r *WorkflowRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workflow, error) {
	var m workflowModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.String()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain()
}

func (r *WorkflowRepo) Update(ctx context.Context, wf *domain.Workflow) error {
	result := r.db.WithContext(ctx).
		Model(&workflowModel{}).
		Where("id = ?", wf.ID.String()).
		Updates(workflowFromDomain(wf))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *WorkflowRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&workflowModel{}, "id = ?", id.String())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *WorkflowRepo) List(ctx context.Context) ([]*domain.Workflow, error) {
	var models []workflowModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Workflow, len(models))
	for i := range models {
		wf, err := models[i].toDomain()
		if err != nil {
			return nil, err
		}
		out[i] = wf
	}
	return out, nil
}

func (r *WorkflowRepo) ListActive(ctx context.Context) ([]*domain.Workflow, error) {
	var models []workflowModel
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Workflow, len(models))
	for i := range models {
		wf, err := models[i].toDomain()
		if err != nil {
			return nil, err
		}
		out[i] = wf
	}
	return out, nil
}
