package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sauravritesh63/GoLang-Project-/internal/domain"
	"github.com/sauravritesh63/GoLang-Project-/internal/repository"
	"gorm.io/gorm"
)

// TaskRepo is a GORM-backed implementation of repository.TaskRepository.
type TaskRepo struct {
	db *gorm.DB
}

// NewTaskRepo constructs a TaskRepo with the supplied *gorm.DB.
func NewTaskRepo(db *gorm.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *domain.Task) error {
	return r.db.WithContext(ctx).Create(taskFromDomain(t)).Error
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	var m taskModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.String()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain()
}

func (r *TaskRepo) Update(ctx context.Context, t *domain.Task) error {
	result := r.db.WithContext(ctx).
		Model(&taskModel{}).
		Where("id = ?", t.ID.String()).
		Updates(taskFromDomain(t))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&taskModel{}, "id = ?", id.String())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *TaskRepo) ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.Task, error) {
	var models []taskModel
	if err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID.String()).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Task, len(models))
	for i := range models {
		t, err := models[i].toDomain()
		if err != nil {
			return nil, err
		}
		out[i] = t
	}
	return out, nil
}
