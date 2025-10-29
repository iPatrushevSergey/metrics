package inmemory

import (
	"sync"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
)

type MemStorageMetricRepository struct {
	mu sync.RWMutex
	DB map[string]model.Metric
}

func NewMemStorageMetricRepository() repository.MetricRepository {
	return &MemStorageMetricRepository{
		DB: make(map[string]model.Metric),
	}
}

func (r *MemStorageMetricRepository) GetByName(name string) (model.Metric, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, exists := r.DB[name]
	return metric, exists
}

func (r *MemStorageMetricRepository) GetAll() map[string]model.Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metricsCopy := make(map[string]model.Metric, len(r.DB))
	for name, metric := range r.DB {
		metricsCopy[name] = metric
	}
	return metricsCopy
}

func (r *MemStorageMetricRepository) Create(name string, metric model.Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DB[name] = metric
}

func (r *MemStorageMetricRepository) Update(name string, metric model.Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DB[name] = metric
}
