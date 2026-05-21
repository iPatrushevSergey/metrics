package usecase

import (
	"context"
	"fmt"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"
)

// GetAllMetrics loads all metrics and builds a sorted list for presentation.
type GetAllMetrics struct {
	metricReader port.MetricReader
	metricSvc    service.MetricService
}

// NewGetAllMetrics returns the get all metrics use case.
func NewGetAllMetrics(metricReader port.MetricReader, metricSvc service.MetricService) port.UseCase[struct{}, []dto.MetricForDisplayOutput] {
	return &GetAllMetrics{metricReader: metricReader, metricSvc: metricSvc}
}

// Execute loads all metrics, sorts by id, formats each value for display.
func (uc *GetAllMetrics) Execute(ctx context.Context, _ struct{}) ([]dto.MetricForDisplayOutput, error) {
	idToMetric, err := uc.metricReader.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}

	metricsWithValue, err := uc.metricSvc.FormatMetricsValue(idToMetric)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}

	metricsForDisplay := make([]dto.MetricForDisplayOutput, len(metricsWithValue))
	for i := range metricsWithValue {
		metricsForDisplay[i] = dto.MetricForDisplayOutput{
			MetricID:    metricsWithValue[i].ID,
			MetricValue: metricsWithValue[i].FormattedValue,
		}
	}
	return metricsForDisplay, nil
}
