package usecase

import (
	"context"
	"fmt"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
)

// RestoreMetricsFromFile loads metrics from the snapshot file into memory.
type RestoreMetricsFromFile struct {
	metricRepo port.MetricRepository
	metricFileRepo port.MetricFileRepository
}

// NewRestoreMetricsFromFile returns the restore use case.
func NewRestoreMetricsFromFile(
	metricRepo port.MetricRepository,
	metricFileRepo port.MetricFileRepository,
) port.UseCase[struct{}, struct{}] {
	return &RestoreMetricsFromFile{metricRepo: metricRepo, metricFileRepo: metricFileRepo}
}

// Execute loads the file snapshot and inserts metrics in one batch.
func (uc *RestoreMetricsFromFile) Execute(ctx context.Context, _ struct{}) (struct{}, error) {
	metrics, err := uc.metricFileRepo.LoadAll(ctx)
	if err != nil {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}
	if len(metrics) == 0 {
		return struct{}{}, nil
	}

	if err := uc.metricRepo.CreateBatchWithParams(ctx, metrics); err != nil {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}
	return struct{}{}, nil
}
