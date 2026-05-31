package port

import (
	"context"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
)

// MetricsGateway is the outbound port to the external metrics server.
type MetricsGateway interface {
	MetricsUpdateBatch(ctx context.Context, metrics []dto.MetricUpdateInput) error
}
