// Package inmemory provides an in-memory implementation of the metric repository.
package inmemory

import (
	"context"
	"sync"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/entity"
)

// MetricMemoryRepository implements a metric repository for in-memory storage.
type MetricMemoryRepository struct {
	mu sync.RWMutex
	DB map[string]entity.Metric
}

// NewMetricMemoryRepository creates a new instance of MetricMemoryRepository.
func NewMetricMemoryRepository() *MetricMemoryRepository {
	return &MetricMemoryRepository{
		DB: make(map[string]entity.Metric),
	}
}

// GetByID returns a metric by id.
func (r *MetricMemoryRepository) GetByID(_ context.Context, id string) (entity.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metric, exists := r.DB[id]
	if !exists {
		return entity.Metric{}, application.ErrNotFound
	}
	return cloneMetric(metric), nil
}

// GetByIDs returns metrics based on a list of IDs.
func (r *MetricMemoryRepository) GetByIDs(_ context.Context, ids []string) (map[string]entity.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]entity.Metric, len(ids))
	for _, id := range ids {
		if metric, exists := r.DB[id]; exists {
			result[id] = cloneMetric(metric)
		}
	}
	return result, nil
}

// GetAll returns all stored metrics.
func (r *MetricMemoryRepository) GetAll(_ context.Context) (map[string]entity.Metric, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metricsCopy := make(map[string]entity.Metric, len(r.DB))
	for name, metric := range r.DB {
		metricsCopy[name] = cloneMetric(metric)
	}
	return metricsCopy, nil
}

// Create inserts a metric.
func (r *MetricMemoryRepository) Create(_ context.Context, metric entity.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.DB[metric.ID]; exists {
		return application.ErrAlreadyExists
	}

	r.DB[metric.ID] = cloneMetric(metric)
	return nil
}

func (r *MetricMemoryRepository) createBatch(_ context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, metric := range metrics {
		if _, exists := r.DB[metric.ID]; exists {
			return application.ErrAlreadyExists
		}
		r.DB[metric.ID] = cloneMetric(metric)
	}

	return nil
}

// CreateBatchWithParams creates multiple metrics.
func (r *MetricMemoryRepository) CreateBatchWithParams(ctx context.Context, metrics []entity.Metric) error {
	return r.createBatch(ctx, metrics)
}

// CreateBatchWithUnnest creates multiple metrics.
func (r *MetricMemoryRepository) CreateBatchWithUnnest(ctx context.Context, metrics []entity.Metric) error {
	return r.createBatch(ctx, metrics)
}

// CreateBatchWithCopy creates multiple metrics.
func (r *MetricMemoryRepository) CreateBatchWithCopy(ctx context.Context, metrics []entity.Metric) error {
	return r.createBatch(ctx, metrics)
}

// CreateBatchWithPrepare creates multiple metrics.
func (r *MetricMemoryRepository) CreateBatchWithPrepare(ctx context.Context, metrics []entity.Metric) error {
	return r.createBatch(ctx, metrics)
}

// Update replaces an existing metric.
func (r *MetricMemoryRepository) Update(_ context.Context, metric entity.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := metric.ID
	if _, exists := r.DB[id]; !exists {
		return application.ErrNotFound
	}

	r.DB[id] = cloneMetric(metric)
	return nil
}

func (r *MetricMemoryRepository) updateBatch(_ context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, metric := range metrics {
		if _, exists := r.DB[metric.ID]; !exists {
			return application.ErrNotFound
		}
		r.DB[metric.ID] = cloneMetric(metric)
	}

	return nil
}

// UpdateBatchWithParams updates multiple metrics.
func (r *MetricMemoryRepository) UpdateBatchWithParams(ctx context.Context, metrics []entity.Metric) error {
	return r.updateBatch(ctx, metrics)
}

// UpdateBatchWithUnnest updates multiple metrics.
func (r *MetricMemoryRepository) UpdateBatchWithUnnest(ctx context.Context, metrics []entity.Metric) error {
	return r.updateBatch(ctx, metrics)
}

// UpdateBatchWithCopy updates multiple metrics.
func (r *MetricMemoryRepository) UpdateBatchWithCopy(ctx context.Context, metrics []entity.Metric) error {
	return r.updateBatch(ctx, metrics)
}

// UpdateBatchWithPrepare updates multiple metrics.
func (r *MetricMemoryRepository) UpdateBatchWithPrepare(ctx context.Context, metrics []entity.Metric) error {
	return r.updateBatch(ctx, metrics)
}

// Ping reports readiness of the storage.
func (r *MetricMemoryRepository) Ping(_ context.Context) error {
	return nil
}

// cloneMetric clones a metric.
func cloneMetric(m entity.Metric) entity.Metric {
	out := entity.Metric{
		ID:    m.ID,
		MType: m.MType,
		Hash:  m.Hash,
	}
	if m.Delta != nil {
		d := *m.Delta
		out.Delta = &d
	}
	if m.Value != nil {
		v := *m.Value
		out.Value = &v
	}
	return out
}
