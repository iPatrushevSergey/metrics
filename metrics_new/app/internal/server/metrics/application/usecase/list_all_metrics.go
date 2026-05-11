package usecase

import (
	"context"
	"fmt"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
)

// ListAllMetrics loads all metrics.
type ListAllMetrics struct {
	metricReader port.MetricReader
}

// NewListAllMetrics returns the list all metrics use case.
func NewListAllMetrics(metricReader port.MetricReader) port.UseCase[struct{}, map[string]dto.MetricOutput] {
	return &ListAllMetrics{metricReader: metricReader}
}

// Execute loads all metrics and maps them keyed by metric id.
func (uc *ListAllMetrics) Execute(ctx context.Context, _ struct{}) (map[string]dto.MetricOutput, error) {
	allMetrics, err := uc.metricReader.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}

	nameToMetric := make(map[string]dto.MetricOutput, len(allMetrics))
	for id, metric := range allMetrics {
		nameToMetric[id] = dto.MetricOutput{
			ID:    metric.ID,
			MType: string(metric.MType),
			Delta: metric.Delta,
			Value: metric.Value,
		}
	}
	return nameToMetric, nil
}
