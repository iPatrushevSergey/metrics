package repository

import "github.com/iPatrushevSergey/metrics/internal/model"

type MetricRepository interface {
	GetByName(name string) (model.Metric, bool)
	GetAll() map[string]model.Metric
	Update(name string, metric model.Metric)
	Create(name string, metric model.Metric)
}
