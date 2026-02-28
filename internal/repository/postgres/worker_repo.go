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

// WorkerRepo is a GORM-backed implementation of repository.WorkerRepository.
type WorkerRepo struct {
	db *gorm.DB
}

// NewWorkerRepo constructs a WorkerRepo with the supplied *gorm.DB.
func NewWorkerRepo(db *gorm.DB) *WorkerRepo {
	return &WorkerRepo{db: db}
}

func (r *WorkerRepo) Create(ctx context.Context, w *domain.Worker) error {
	return r.db.WithContext(ctx).Create(workerFromDomain(w)).Error
}

func (r *WorkerRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Worker, error) {
	var m workerModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.String()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain()
}

func (r *WorkerRepo) Update(ctx context.Context, w *domain.Worker) error {
	result := r.db.WithContext(ctx).
		Model(&workerModel{}).
		Where("id = ?", w.ID.String()).
		Updates(workerFromDomain(w))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *WorkerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&workerModel{}, "id = ?", id.String())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *WorkerRepo) ListActive(ctx context.Context) ([]*domain.Worker, error) {
	var models []workerModel
	if err := r.db.WithContext(ctx).
		Where("status = ?", string(domain.WorkerStatusActive)).
		Order("last_heartbeat DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Worker, len(models))
	for i := range models {
		w, err := models[i].toDomain()
		if err != nil {
			return nil, err
		}
		out[i] = w
	}
	return out, nil
}

func (r *WorkerRepo) UpdateHeartbeat(ctx context.Context, id uuid.UUID, at time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&workerModel{}).
		Where("id = ?", id.String()).
		Update("last_heartbeat", at)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
