package inmemory

import (
	"sync"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

type MetricRepository interface {
	GetByName(name string) (model.Metric, bool)
	Update(name string, metric model.Metric)
	Create(name string, metric model.Metric)
}

type MemStorageMetricRepository struct {
	mu sync.RWMutex
	DB map[string]model.Metric
}

func NewMemStorageMetricRepository() MetricRepository {
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
