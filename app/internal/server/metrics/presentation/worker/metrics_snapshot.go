// Package worker runs background tasks.
package worker

import (
	"context"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/factory"
)

// SnapshotWorker periodically persists the in-memory metrics snapshot to a file.
type SnapshotWorker struct {
	useCases      factory.UseCaseFactory
	log           port.Logger
	storeInterval time.Duration
}

// NewSnapshotWorker creates a metrics snapshot background worker.
func NewSnapshotWorker(
	useCases factory.UseCaseFactory,
	log port.Logger,
	storeInterval time.Duration,
) *SnapshotWorker {
	return &SnapshotWorker{
		useCases:      useCases,
		log:           log,
		storeInterval: storeInterval,
	}
}

// Run executes the worker loop.
func (w *SnapshotWorker) Run(ctx context.Context) {
	w.log.Info("metrics snapshot worker started", "store_interval", w.storeInterval)

	ticker := time.NewTicker(w.storeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("metrics snapshot worker stopped")
			return
		case <-ticker.C:
			count, err := w.useCases.MetricsSnapshotUseCase().Execute(ctx, struct{}{})
			if err != nil {
				w.log.Error("metrics snapshot failed", "error", err)
				continue
			}
			w.log.Debug("metrics snapshot saved", "count", count)
		}
	}
}
