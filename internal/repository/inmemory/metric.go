package inmemory

import (
	"context"
	"sync"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
)

// MemStorageMetricRepository implements a MetricRepository for in-memory storage
type MemStorageMetricRepository struct {
	mu sync.RWMutex
	DB map[string]model.Metric
}

// NewMemStorageMetricRepository creates a new instance of MemStorageMetricRepository
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

// GetByIDs returns metrics based on a list of IDs
func (r *MemStorageMetricRepository) GetByIDs(ctx context.Context, ids []string) (map[string]model.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]model.Metric, len(ids))
	for _, id := range ids {
		if metric, exists := r.DB[id]; exists {
			result[id] = metric
		}
	}
	return result, nil
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

// CreateBatch creates multiple metrics in one call under one lock
func (r *MemStorageMetricRepository) CreateBatch(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, metric := range metrics {
		if _, exists := r.DB[metric.ID]; exists {
			return repository.ErrAlreadyExists
		}
		r.DB[metric.ID] = metric
	}

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

// UpdateBatch updates multiple metrics in one call under one lock
func (r *MemStorageMetricRepository) UpdateBatch(ctx context.Context, metrics []model.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, metric := range metrics {
		if _, exists := r.DB[metric.ID]; !exists {
			return repository.ErrNotFound
		}
		r.DB[metric.ID] = metric
	}

	return nil
}

func (r *MemStorageMetricRepository) Ping(ctx context.Context) error {
	return nil
}
