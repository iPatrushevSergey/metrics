package repository

import (
	"context"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

type MetricRepository interface {
	GetByID(ctx context.Context, id string) (model.Metric, bool)
	GetAll(ctx context.Context) map[string]model.Metric
	Update(ctx context.Context, id string, metric model.Metric) error
	Create(ctx context.Context, metric model.Metric) error
	Ping(ctx context.Context) error
}
