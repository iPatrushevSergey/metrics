package worker

import (
	"context"
	"sync"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/presentation/factory"
)

// ReportWorker runs ReportBatchTick on each interval.
type ReportWorker struct {
	sendCtx  context.Context
	useCases factory.UseCaseFactory
	log      port.Logger
	interval time.Duration
	sendsWg  sync.WaitGroup
}

// NewReportWorker creates a report background worker.
func NewReportWorker(
	sendCtx context.Context,
	useCases factory.UseCaseFactory,
	log port.Logger,
	interval time.Duration,
) *ReportWorker {
	return &ReportWorker{
		sendCtx:  sendCtx,
		useCases: useCases,
		log:      log,
		interval: interval,
	}
}

// Wait blocks until all in-flight report sends finish.
func (w *ReportWorker) Wait() {
	w.sendsWg.Wait()
}

// Run executes the worker loop.
func (w *ReportWorker) Run(ctx context.Context) {
	w.log.Info("report worker started", "interval", w.interval.String())

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("report worker stopped")
			return
		case <-ticker.C:
			w.log.Debug("report tick")
			w.sendsWg.Add(1)
			go func() {
				defer w.sendsWg.Done()
				if _, err := w.useCases.ReportBatchTick().Execute(w.sendCtx, struct{}{}); err != nil {
					w.log.Error("report tick failed", "error", err)
				}
			}()
		}
	}
}
