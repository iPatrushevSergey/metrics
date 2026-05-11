package usecase

import (
	"context"
	"fmt"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
)

// GetAllMetrics loads all metrics.
type GetAllMetrics struct {
	metricReader port.MetricReader
}

// NewGetAllMetrics returns the get all metrics use case.
func NewGetAllMetrics(metricReader port.MetricReader) port.UseCase[struct{}, map[string]dto.MetricOutput] {
	return &GetAllMetrics{metricReader: metricReader}
}

// Execute loads all metrics and maps them keyed by ID.
func (uc *GetAllMetrics) Execute(ctx context.Context, _ struct{}) (map[string]dto.MetricOutput, error) {
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
