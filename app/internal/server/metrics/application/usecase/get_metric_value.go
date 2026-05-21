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

// GetMetricValue loads a single metric value.
type GetMetricValue struct {
	metricReader port.MetricReader
	metricSvc    service.MetricService
}

// NewGetMetricValue returns the get metric value use case.
func NewGetMetricValue(metricReader port.MetricReader, metricSvc service.MetricService) port.UseCase[dto.GetMetricValueInput, string] {
	return &GetMetricValue{metricReader: metricReader, metricSvc: metricSvc}
}

// Execute loads the metric, checks type matches path, and formats the value.
func (uc *GetMetricValue) Execute(ctx context.Context, inDTO dto.GetMetricValueInput) (string, error) {
	mType, err := uc.metricSvc.CheckMetricType(inDTO.MType)
	if err != nil {
		return "", application.ErrBadMetricType
	}

	metric, err := uc.metricReader.GetByID(ctx, inDTO.ID)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrNotFound):
			return "", application.ErrNotFound
		default:
			return "", fmt.Errorf("%w: %v", application.ErrInternal, err)
		}
	}
	if err := metric.MatchMetricTypes(mType); err != nil {
		switch {
		case errors.Is(err, entity.ErrMetricTypeMismatch):
			return "", application.ErrNotFound
		default:
			return "", err
		}
	}

	value, err := metric.FormatValueAsString()
	if err != nil {
		return "", fmt.Errorf("%w: %v", application.ErrInternal, err)
	}
	return value, nil
}
