// Package lifecycle manages the application lifecycle.
package lifecycle

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/factory"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/worker"
)

// App represents the application lifecycle.
type App struct {
	Server          *http.Server
	UseCases        factory.UseCaseFactory
	Log             port.Logger
	ShutdownTimeout time.Duration

	FileStorage   bool
	Restore       bool
	StoreInterval time.Duration

	AuditPublisher    port.AuditPublisher
	AuditFileRepo     port.AuditFileRepository
	AuditFileEvents   <-chan dto.AuditEvent
	AuditRemoteEvents <-chan dto.AuditEvent

	cancelSnapshotWorker context.CancelFunc
	cancelAuditWorker    context.CancelFunc
}

// Start starts the application.
func (a *App) Start() {
	ctx := context.Background()

	if a.FileStorage && a.Restore {
		if uc := a.UseCases.RestoreMetricsFromFileUseCase(); uc != nil {
			if _, err := uc.Execute(ctx, struct{}{}); err != nil {
				a.Log.Warn("restore metrics from file failed", "error", err)
			}
		}
	}

	if a.AuditPublisher != nil {
		workerCtx, cancel := context.WithCancel(ctx)
		a.cancelAuditWorker = cancel
		if a.AuditFileEvents != nil {
			go worker.NewAuditFileSubscriber(a.AuditFileEvents, a.UseCases, a.Log).Run(workerCtx)
		}
		if a.AuditRemoteEvents != nil {
			go worker.NewAuditRemoteSubscriber(a.AuditRemoteEvents, a.UseCases, a.Log).Run(workerCtx)
		}
	}

	if a.FileStorage && a.StoreInterval > 0 {
		workerCtx, cancel := context.WithCancel(ctx)
		a.cancelSnapshotWorker = cancel
		go worker.NewSnapshotWorker(a.UseCases, a.Log, a.StoreInterval).Run(workerCtx)
	}

	go func() {
		a.Log.Info("server listening", "address", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Log.Error("server failed", "error", err)
		}
	}()
}

// Stop stops the application.
func (a *App) Stop() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	a.Log.Info("shutdown signal received, stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), a.ShutdownTimeout)
	defer cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		a.Log.Error("server shutdown failed", "error", err)
		return err
	}

	if a.cancelSnapshotWorker != nil {
		a.cancelSnapshotWorker()
	}

	if a.FileStorage {
		if uc := a.UseCases.MetricsSnapshotUseCase(); uc != nil {
			if _, err := uc.Execute(ctx, struct{}{}); err != nil {
				a.Log.Error("metrics snapshot on shutdown failed", "error", err)
			}
		}
	}

	if a.AuditPublisher != nil {
		if a.cancelAuditWorker != nil {
			a.cancelAuditWorker()
		}
		a.AuditPublisher.Unsubscribe("file")
		a.AuditPublisher.Unsubscribe("remote")
		if a.AuditFileRepo != nil {
			if err := a.AuditFileRepo.Close(); err != nil {
				a.Log.Error("audit file repository close failed", "error", err)
			}
		}
		if err := a.AuditPublisher.Close(ctx); err != nil {
			a.Log.Error("audit shutdown failed", "error", err)
		}
	}

	a.Log.Info("server stopped gracefully")
	return nil
}
