// Package worker runs periodic collector tasks.
package worker

import (
	"context"
	"errors"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/presentation/factory"
)

// PollRuntimeWorker runs PollRuntimeTick on each interval.
type PollRuntimeWorker struct {
	useCases factory.UseCaseFactory
	log      port.Logger
	interval time.Duration
}

// NewPollRuntimeWorker creates a poll runtime background worker.
func NewPollRuntimeWorker(useCases factory.UseCaseFactory, log port.Logger, interval time.Duration) *PollRuntimeWorker {
	return &PollRuntimeWorker{useCases: useCases, log: log, interval: interval}
}

// Run executes the worker loop.
func (w *PollRuntimeWorker) Run(ctx context.Context) {
	w.log.Info("poll runtime worker started", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("poll runtime worker stopped")
			return
		case <-ticker.C:
			if _, err := w.useCases.PollRuntimeTick().Execute(ctx, struct{}{}); err != nil && !errors.Is(err, context.Canceled) {
				w.log.Error("poll runtime tick failed", "error", err)
			}
		}
	}
}
