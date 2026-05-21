package port

import (
	"context"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
)

// MetricFileRepository persists metrics snapshot to a file.
type MetricFileRepository interface {
	SaveAll(ctx context.Context, metrics map[string]entity.Metric) error
	LoadAll(ctx context.Context) ([]entity.Metric, error)
}
