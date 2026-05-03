package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/domain/service"
)

// ReportBatchTick sends metrics batch built from the current metrics repository.
type ReportBatchTick struct {
	metricsRepo   port.MetricsRepository
	metricsClient port.MetricsClient
	log           port.Logger
	interval      time.Duration
}

// NewReportBatchTick returns report batch tick use case.
func NewReportBatchTick(
	metricsRepo port.MetricsRepository,
	metricsClient port.MetricsClient,
	log port.Logger,
	reportInterval time.Duration,
) *ReportBatchTick {
	return &ReportBatchTick{
		metricsRepo:   metricsRepo,
		metricsClient: metricsClient,
		log:           log,
		interval:      reportInterval,
	}
}

// Run builds a batch from the metrics repository and sends it to the metrics server.
func (uc *ReportBatchTick) Run(ctx context.Context) (int, error) {
	start := time.Now()
	defer func() {
		d := time.Since(start)
		if d > uc.interval {
			uc.log.Warn("metrics send slower than report interval",
				"duration", d.String(),
				"report_interval", uc.interval.String(),
			)
		}
	}()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	snap := uc.metricsRepo.GetSystemMetrics()
	gauges := service.GaugesFromState(&snap.Runtime.MemStats, &snap.Poll, &snap.Gopsutil)
	counters := service.CountersFromState(&snap.Poll)

	total := len(gauges) + len(counters)
	if total == 0 {
		uc.log.Debug("no metrics to send")
		return 0, nil
	}

	out := make([]dto.MetricUpdateInput, 0, total)
	for name, v := range gauges {
		val := v
		out = append(out, dto.MetricUpdateInput{ID: name, MType: "gauge", Value: &val})
	}
	for name, d := range counters {
		delta := d
		out = append(out, dto.MetricUpdateInput{ID: name, MType: "counter", Delta: &delta})
	}

	if err := uc.metricsClient.MetricsUpdateBatch(ctx, out); err != nil {
		if errors.Is(err, context.Canceled) {
			uc.log.Info("sending metrics batch canceled")
			return 0, err
		}
		return 0, err
	}
	uc.log.Debug("batch sent", "count", len(out))
	return len(out), nil
}
