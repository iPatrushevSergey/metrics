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

// UpdateMetric creates or updates a metric.
type UpdateMetric struct {
	metricRepo port.MetricRepository
	metricSvc  service.MetricService
}

// NewUpdateMetric returns the update metric use case.
func NewUpdateMetric(metricRepo port.MetricRepository, metricSvc service.MetricService) port.UseCase[dto.UpdateMetricInput, struct{}] {
	return &UpdateMetric{metricRepo: metricRepo, metricSvc: metricSvc}
}

// Execute loads the existing metric if any, validates the input, merges the values, creates or updates the metric.
func (uc *UpdateMetric) Execute(ctx context.Context, inDTO dto.UpdateMetricInput) (struct{}, error) {
	mType, err := uc.metricSvc.CheckMetricType(inDTO.MType)
	if err != nil {
		return struct{}{}, application.ErrBadMetricType
	}

	newMetric, err := entity.NewMetric(mType, inDTO.ID, inDTO.Value)
	if err != nil {
		return struct{}{}, application.ErrBadMetricValue
	}

	existingMetric, err := uc.metricRepo.GetByID(ctx, newMetric.ID)
	if err != nil && !errors.Is(err, application.ErrNotFound) {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}

	if errors.Is(err, application.ErrNotFound) {
		if err := uc.metricRepo.Create(ctx, newMetric); err != nil {
			return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
		}
		return struct{}{}, nil
	}

	if err := existingMetric.MatchMetricTypes(newMetric.MType); err != nil {
		switch {
		case errors.Is(err, entity.ErrMetricTypeMismatch):
			return struct{}{}, application.ErrNotFound
		default:
			return struct{}{}, err
		}
	}

	if err := existingMetric.ApplyUpdate(newMetric); err != nil {
		switch {
		case errors.Is(err, entity.ErrUnsupportedMetricType):
			return struct{}{}, application.ErrBadMetricType
		default:
			return struct{}{}, err
		}
	}

	if err := uc.metricRepo.Update(ctx, existingMetric); err != nil {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}
	return struct{}{}, nil
}
