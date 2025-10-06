package inmemory

import "github.com/iPatrushevSergey/metrics/internal/model"

type MemStorageMetricRepository struct {
	metrics map[string]model.Metric
}

func NewMemStorageMetricRepository() *MemStorageMetricRepository {
	return &MemStorageMetricRepository{
		metrics: make(map[string]model.Metric),
	}
}

func (r *MemStorageMetricRepository) GetByName(name string) (model.Metric, bool) {
	metric, exists := r.metrics[name]
	return metric, exists
}

func (r *MemStorageMetricRepository) Create(name string, metric model.Metric) {
	r.metrics[name] = metric
}

func (r *MemStorageMetricRepository) Update(name string, metric model.Metric) {
	r.metrics[name] = metric
}
