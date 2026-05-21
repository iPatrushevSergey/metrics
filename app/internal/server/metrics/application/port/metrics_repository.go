package port

import (
	"context"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/entity"
)

// MetricReader provides read-only access to metrics for metrics module.
type MetricReader interface {
	GetByID(ctx context.Context, id string) (entity.Metric, error)
	GetByIDs(ctx context.Context, ids []string) (map[string]entity.Metric, error)
	GetAll(ctx context.Context) (map[string]entity.Metric, error)
	Ping(ctx context.Context) error
}

// MetricWriter provides write access to metrics for metrics module.
type MetricWriter interface {
	Create(ctx context.Context, metric entity.Metric) error
	CreateBatchWithPrepare(ctx context.Context, metrics []entity.Metric) error
	CreateBatchWithParams(ctx context.Context, metrics []entity.Metric) error
	CreateBatchWithUnnest(ctx context.Context, metrics []entity.Metric) error
	CreateBatchWithCopy(ctx context.Context, metrics []entity.Metric) error
	Update(ctx context.Context, metric entity.Metric) error
	UpdateBatchWithPrepare(ctx context.Context, metrics []entity.Metric) error
	UpdateBatchWithParams(ctx context.Context, metrics []entity.Metric) error
	UpdateBatchWithUnnest(ctx context.Context, metrics []entity.Metric) error
	UpdateBatchWithCopy(ctx context.Context, metrics []entity.Metric) error
}

// MetricRepository combines reader and writer for metrics DI wiring.
type MetricRepository interface {
	MetricReader
	MetricWriter
}
