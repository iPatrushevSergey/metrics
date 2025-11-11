package repository

import "github.com/iPatrushevSergey/metrics/internal/model"

type MetricRepository interface {
	GetByID(id string) (model.Metric, bool)
	GetAll() map[string]model.Metric
	Update(id string, metric model.Metric)
	Create(metric model.Metric)
}
