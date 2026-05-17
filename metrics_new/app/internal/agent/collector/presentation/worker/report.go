package worker

import (
	"context"
	"sync"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/presentation/factory"
)

// ReportLoop runs the report use case on each tick.
type ReportLoop struct {
	sendCtx  context.Context
	useCases factory.UseCaseFactory
	log      port.Logger
	interval time.Duration
	sendsWg  sync.WaitGroup
}

// NewReportLoop initializes the report loop.
func NewReportLoop(
	sendCtx context.Context,
	useCases factory.UseCaseFactory,
	log port.Logger,
	interval time.Duration,
) *ReportLoop {
	return &ReportLoop{
		sendCtx:  sendCtx,
		useCases: useCases,
		log:      log,
		interval: interval,
	}
}

// WaitSendsWg waits until all report goroutines finish.
func (w *ReportLoop) WaitSendsWg() {
	w.sendsWg.Wait()
}

// RunReportTickerLoop runs the report use case on each tick until pollCtx is canceled.
func (w *ReportLoop) RunReportTickerLoop(pollCtx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.log.Info("report worker started", "interval", w.interval.String())
	for {
		select {
		case <-pollCtx.Done():
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
