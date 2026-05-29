// Package lifecycle manages the agent application lifecycle.
package bootstrap

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/adapters/metrics_gateway"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/presentation/factory"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/presentation/worker"
)

// App represents the application lifecycle.
type App struct {
	UseCases       factory.UseCaseFactory
	Log            port.Logger
	PollInterval   time.Duration
	ReportInterval time.Duration

	PollCtx context.Context
	SendCtx context.Context

	CancelPoll context.CancelFunc
	CancelSend context.CancelFunc

	BufferedGateway *metricsgateway.BufferedMetricsGateway
	GatewayCloser   io.Closer

	reportWorker *worker.ReportWorker
}

// Start starts the application.
func (a *App) Start() {
	if a.BufferedGateway != nil {
		a.BufferedGateway.Start()
	}

	go worker.NewPollRuntimeWorker(a.UseCases, a.Log, a.PollInterval).Run(a.PollCtx)
	go worker.NewPollGopsutilWorker(a.UseCases, a.Log, a.PollInterval).Run(a.PollCtx)

	a.reportWorker = worker.NewReportWorker(a.SendCtx, a.UseCases, a.Log, a.ReportInterval)
	go a.reportWorker.Run(a.PollCtx)
}

// Stop stops the application.
func (a *App) Stop() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	a.Log.Info("shutdown signal received, stopping poll loops...")
	if a.CancelPoll != nil {
		a.CancelPoll()
	}

	if a.reportWorker != nil {
		a.reportWorker.Wait()
	}

	if a.BufferedGateway != nil {
		a.BufferedGateway.Stop()
	}

	if a.CancelSend != nil {
		a.CancelSend()
	}

	if a.GatewayCloser != nil {
		if err := a.GatewayCloser.Close(); err != nil {
			a.Log.Error("metrics gateway close failed", "error", err)
		}
	}

	a.Log.Info("agent stopped gracefully")
	return nil
}
