package usecase

import (
	"context"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
)

// PollGopsutilTick records host metrics sample into the metrics repository.
type PollGopsutilTick struct {
	metricsSampler port.MetricsSampler
	metricsRepo    port.MetricsRepository
	log            port.Logger
}

// NewPollGopsutilTick returns poll gopsutil tick use case.
func NewPollGopsutilTick(
	metricsSampler port.MetricsSampler,
	metricsRepo port.MetricsRepository,
	log port.Logger,
) *PollGopsutilTick {
	return &PollGopsutilTick{
		metricsSampler: metricsSampler,
		metricsRepo:    metricsRepo,
		log:            log,
	}
}

// Run performs a single gopsutil poll tick.
func (uc *PollGopsutilTick) Run(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	total, free, err := uc.metricsSampler.ReadVirtualMemory()
	if err != nil {
		uc.log.Error("memory metrics", "error", err)
		return 0, nil
	}
	percents, err := uc.metricsSampler.ReadCPUPercent(time.Second, true)
	if err != nil {
		uc.log.Error("cpu metrics", "error", err)
		percents = nil
	}

	uc.metricsRepo.UpdateGopsutilMetrics(float64(total), float64(free), percents)
	return 1, nil
}
