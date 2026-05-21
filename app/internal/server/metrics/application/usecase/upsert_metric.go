package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"
)

// UpsertMetric creates or updates a metric.
type UpsertMetric struct {
	metricRepo     port.MetricRepository
	metricFileRepo port.MetricFileRepository
	auditPublisher port.AuditPublisher
	metricSvc      service.MetricService
}

// NewUpsertMetric returns the upsert metric use case.
func NewUpsertMetric(
	metricRepo port.MetricRepository,
	metricSvc service.MetricService,
	metricFileRepo port.MetricFileRepository,
	auditPublisher port.AuditPublisher,
) port.UseCase[dto.UpsertMetricInput, struct{}] {
	return &UpsertMetric{
		metricRepo:     metricRepo,
		metricFileRepo: metricFileRepo,
		auditPublisher: auditPublisher,
		metricSvc:      metricSvc,
	}
}

// Execute validates input, loads existing row if any, merges, creates or updates.
func (uc *UpsertMetric) Execute(ctx context.Context, inDTO dto.UpsertMetricInput) (struct{}, error) {
	mType, err := uc.metricSvc.CheckMetricType(inDTO.MType)
	if err != nil {
		return struct{}{}, application.ErrBadMetricType
	}

	newMetric := entity.Metric{
		ID:    inDTO.ID,
		MType: mType,
		Delta: inDTO.Delta,
		Value: inDTO.Value,
		Hash:  inDTO.Hash,
	}

	if err := newMetric.ValidateMetricValues(); err != nil {
		switch {
		case errors.Is(err, entity.ErrUnsupportedMetricType):
			return struct{}{}, application.ErrBadMetricType
		case errors.Is(err, entity.ErrMissingCounterDelta), errors.Is(err, entity.ErrMissingGaugeValue):
			return struct{}{}, application.ErrBadMetricValue
		default:
			return struct{}{}, err
		}
	}

	existingMetric, err := uc.metricRepo.GetByID(ctx, newMetric.ID)
	if err != nil && !errors.Is(err, application.ErrNotFound) {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}

	if errors.Is(err, application.ErrNotFound) {
		if err := uc.metricRepo.Create(ctx, newMetric); err != nil {
			return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
		}
		if uc.metricFileRepo != nil {
			metrics, err := uc.metricRepo.GetAll(ctx)
			if err != nil {
				return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
			}
			if err := uc.metricFileRepo.SaveAll(ctx, metrics); err != nil {
				return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
			}
		}
		if uc.auditPublisher != nil {
			uc.auditPublisher.Publish(dto.AuditEvent{
				TS:        time.Now().Unix(),
				Metrics:   []string{inDTO.ID},
				IPAddress: inDTO.IPAddress,
			})
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
	if uc.metricFileRepo != nil {
		metrics, err := uc.metricRepo.GetAll(ctx)
		if err != nil {
			return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
		}
		if err := uc.metricFileRepo.SaveAll(ctx, metrics); err != nil {
			return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
		}
	}
	if uc.auditPublisher != nil {
		uc.auditPublisher.Publish(dto.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   []string{inDTO.ID},
			IPAddress: inDTO.IPAddress,
		})
	}
	return struct{}{}, nil
}
