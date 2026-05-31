package usecase

import (
	"context"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
)

// PollRuntimeTick records runtime + custom poll tick into the metrics repository.
type PollRuntimeTick struct {
	metricsSampler port.MetricsSampler
	metricsRepo    port.MetricsRepository
	randGenerator  func() float64
}

// NewPollRuntimeTick returns poll runtime tick use case.
func NewPollRuntimeTick(
	metricsSampler port.MetricsSampler,
	metricsRepo port.MetricsRepository,
	randGenerator func() float64,
) port.UseCase[struct{}, int] {
	return &PollRuntimeTick{
		metricsSampler: metricsSampler,
		metricsRepo:    metricsRepo,
		randGenerator:  randGenerator,
	}
}

// Execute performs a single poll tick.
func (uc *PollRuntimeTick) Execute(ctx context.Context, _ struct{}) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	uc.metricsRepo.UpdateRuntimeMetrics(uc.metricsSampler.ReadMemStats(), uc.randGenerator())
	return 1, nil
}
