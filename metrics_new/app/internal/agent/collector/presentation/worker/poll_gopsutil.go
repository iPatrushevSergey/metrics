package worker

import (
	"context"
	"errors"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/presentation/factory"
)

// PollGopsutilWorker runs PollGopsutilTick on each interval.
type PollGopsutilWorker struct {
	useCases factory.UseCaseFactory
	log      port.Logger
	interval time.Duration
}

// NewPollGopsutilWorker creates a poll gopsutil background worker.
func NewPollGopsutilWorker(useCases factory.UseCaseFactory, log port.Logger, interval time.Duration) *PollGopsutilWorker {
	return &PollGopsutilWorker{useCases: useCases, log: log, interval: interval}
}

// Run executes the worker loop.
func (w *PollGopsutilWorker) Run(ctx context.Context) {
	w.log.Info("poll gopsutil worker started", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("poll gopsutil worker stopped")
			return
		case <-ticker.C:
			if _, err := w.useCases.PollGopsutilTick().Execute(ctx, struct{}{}); err != nil && !errors.Is(err, context.Canceled) {
				w.log.Error("poll gopsutil tick failed", "error", err)
			}
		}
	}
}
