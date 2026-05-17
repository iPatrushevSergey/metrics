package usecase

import (
	"context"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
)

// MetricsSnapshot loads all metrics from memory and rewrites the snapshot file.
type MetricsSnapshot struct {
	metricReader   port.MetricReader
	metricFileRepo port.MetricFileRepository
}

// NewMetricsSnapshot returns a use case that flushes metrics to disk.
func NewMetricsSnapshot(
	metricReader port.MetricReader,
	metricFileRepo port.MetricFileRepository,
) port.UseCase[struct{}, int] {
	return &MetricsSnapshot{metricReader: metricReader, metricFileRepo: metricFileRepo}
}

// Execute writes the full metrics snapshot to the file.
func (uc *MetricsSnapshot) Execute(ctx context.Context, _ struct{}) (int, error) {
	metrics, err := uc.metricReader.GetAll(ctx)
	if err != nil {
		return 0, err
	}
	if err := uc.metricFileRepo.SaveAll(ctx, metrics); err != nil {
		return 0, err
	}
	return len(metrics), nil
}
