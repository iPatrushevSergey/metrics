package repository

import (
	"context"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

type MetricRepository interface {
	GetByID(id string) (model.Metric, bool)
	GetAll() map[string]model.Metric
	Update(id string, metric model.Metric) error
	Create(metric model.Metric) error
	Ping(ctx context.Context) error
}
