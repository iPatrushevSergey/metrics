package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/entity"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/service"
)

// UpsertMetricsBatch creates or updates a batch of metrics.
type UpsertMetricsBatch struct {
	metricRepo port.MetricRepository
	metricSvc  service.MetricService
	transactor port.Transactor
}

// NewUpsertMetricsBatch returns the batch upsert use case.
func NewUpsertMetricsBatch(
	metricRepo port.MetricRepository,
	metricSvc service.MetricService,
	transactor port.Transactor,
) port.UseCase[dto.UpsertMetricsBatchInput, struct{}] {
	return &UpsertMetricsBatch{metricRepo: metricRepo, metricSvc: metricSvc, transactor: transactor}
}

// Execute validates each metric, folds duplicate ids within the request preserving order, then persists.
func (uc *UpsertMetricsBatch) Execute(ctx context.Context, inDTO dto.UpsertMetricsBatchInput) (struct{}, error) {
	if len(inDTO.Metrics) == 0 {
		return struct{}{}, nil
	}

	metrics := make([]entity.Metric, 0, len(inDTO.Metrics))
	for _, rawMetric := range inDTO.Metrics {
		mType, err := uc.metricSvc.CheckMetricType(strings.TrimSpace(rawMetric.MType))
		if err != nil {
			return struct{}{}, application.ErrBadMetricType
		}

		newMetric := entity.Metric{
			ID:    rawMetric.ID,
			MType: mType,
			Delta: rawMetric.Delta,
			Value: rawMetric.Value,
			Hash:  rawMetric.Hash,
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
		metrics = append(metrics, newMetric)
	}

	mergedMetrics, err := uc.metricSvc.MergeMetricsByID(metrics)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrMetricTypeMismatch):
			return struct{}{}, application.ErrNotFound
		case errors.Is(err, entity.ErrUnsupportedMetricType):
			return struct{}{}, application.ErrBadMetricType
		default:
			return struct{}{}, err
		}
	}

	metricIDs := uc.metricSvc.CollectIDs(mergedMetrics)

	existingIDToMetric, err := uc.metricRepo.GetByIDs(ctx, metricIDs)
	if err != nil {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}

	createList, updateList, err := uc.metricSvc.BuildCreateUpdateBatches(existingIDToMetric, mergedMetrics)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrMetricTypeMismatch):
			return struct{}{}, application.ErrNotFound
		case errors.Is(err, entity.ErrUnsupportedMetricType):
			return struct{}{}, application.ErrBadMetricType
		default:
			return struct{}{}, err
		}
	}

	if len(createList) == 0 && len(updateList) == 0 {
		return struct{}{}, nil
	}

	err = uc.transactor.RunInTransaction(ctx, func(txCtx context.Context) error {
		if len(createList) > 0 {
			if err := uc.metricRepo.CreateBatchWithUnnest(txCtx, createList); err != nil {
				return err
			}
		}
		if len(updateList) > 0 {
			if err := uc.metricRepo.UpdateBatchWithUnnest(txCtx, updateList); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}

	return struct{}{}, nil
}
