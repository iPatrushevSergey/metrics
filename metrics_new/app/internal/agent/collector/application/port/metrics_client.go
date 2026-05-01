package port

import (
	"context"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/dto"
)

// MetricsSender communicates with the external metrics server.
type MetricsClient interface {
	UpdateMetricsBatch(ctx context.Context, metrics []dto.MetricUpdateInput) error
}
