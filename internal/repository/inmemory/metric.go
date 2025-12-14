package inmemory

import (
	"context"
	"sync"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
)

// MemStorageMetricRepository реализует MetricRepository для хранения в памяти
type MemStorageMetricRepository struct {
	mu sync.RWMutex
	DB map[string]model.Metric
}

// NewMemStorageMetricRepository создает новый экземпляр MemStorageMetricRepository
func NewMemStorageMetricRepository() repository.MetricRepository {
	return &MemStorageMetricRepository{
		DB: make(map[string]model.Metric),
	}
}

func (r *MemStorageMetricRepository) GetByID(ctx context.Context, id string) (model.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metric, exists := r.DB[id]
	if !exists {
		return model.Metric{}, repository.ErrNotFound
	}
	return metric, nil
}

func (r *MemStorageMetricRepository) GetAll(ctx context.Context) (map[string]model.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metricsCopy := make(map[string]model.Metric, len(r.DB))
	for name, metric := range r.DB {
		metricsCopy[name] = metric
	}
	return metricsCopy, nil
}

func (r *MemStorageMetricRepository) Create(ctx context.Context, metric model.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.DB[metric.ID]; exists {
		return repository.ErrAlreadyExists
	}

	r.DB[metric.ID] = metric
	return nil
}

func (r *MemStorageMetricRepository) Update(ctx context.Context, id string, metric model.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.DB[id]; !exists {
		return repository.ErrNotFound
	}

	r.DB[id] = metric
	return nil
}

func (r *MemStorageMetricRepository) Ping(ctx context.Context) error {
	return nil
}
