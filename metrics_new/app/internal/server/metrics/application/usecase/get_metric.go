package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/entity"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/service"
)

// GetMetric loads a single metric.
type GetMetric struct {
	metricReader port.MetricReader
	metricSvc    service.MetricService
}

// NewGetMetric returns the get metric use case.
func NewGetMetric(metricReader port.MetricReader, metricSvc service.MetricService) port.UseCase[dto.GetMetricInput, dto.MetricOutput] {
	return &GetMetric{metricReader: metricReader, metricSvc: metricSvc}
}

// Execute loads the metric, checks type matches, and validates values.
func (uc *GetMetric) Execute(ctx context.Context, inDTO dto.GetMetricInput) (dto.MetricOutput, error) {
	mType, err := uc.metricSvc.CheckMetricType(inDTO.MType)
	if err != nil {
		return dto.MetricOutput{}, application.ErrBadMetricType
	}

	metric, err := uc.metricReader.GetByID(ctx, inDTO.ID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrNotFound):
			return dto.MetricOutput{}, application.ErrNotFound
		default:
			return dto.MetricOutput{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
		}
	}
	if err := metric.MatchMetricTypes(mType); err != nil {
		switch {
		case errors.Is(err, entity.ErrMetricTypeMismatch):
			return dto.MetricOutput{}, application.ErrNotFound
		default:
			return dto.MetricOutput{}, err
		}
	}

	if err := metric.ValidateMetricValues(); err != nil {
		switch {
		case errors.Is(err, entity.ErrUnsupportedMetricType):
			return dto.MetricOutput{}, application.ErrBadMetricType
		case errors.Is(err, entity.ErrMissingCounterDelta), errors.Is(err, entity.ErrMissingGaugeValue):
			return dto.MetricOutput{}, application.ErrBadMetricValue
		default:
			return dto.MetricOutput{}, err
		}
	}

	return dto.MetricOutput{
		ID:    metric.ID,
		MType: string(metric.MType),
		Delta: metric.Delta,
		Value: metric.Value,
	}, nil
}
