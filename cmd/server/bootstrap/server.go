package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/filestorage"
	"github.com/iPatrushevSergey/metrics/internal/logger"
)

const serverShutdownTimeout = 5 * time.Second

// StartServer starts the HTTP server in a goroutine
func StartServer(server *http.Server, loggerInstance logger.Logger) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			loggerInstance.Error("Server failed to start", zap.Error(err))
			errCh <- err
		}
	}()
	return errCh
}

// WaitForShutdown waits for shutdown signal and performs graceful shutdown
func WaitForShutdown(app *App, cfg config.ServerConfig, loggerInstance logger.Logger, serverErrCh <-chan error) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-quit:
			loggerInstance.Debug("The completion signal has been received, starting the stop...")
		case err, ok := <-serverErrCh:
			if !ok {
				serverErrCh = nil
				continue
			}
			return err
		}
		break
	}

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	// Save metrics on shutdown if using file storage
	if cfg.DatabaseDSN == "" && app.FileStorage != nil {
		if app.PeriodicSaver != nil {
			loggerInstance.Debug("Stopping periodic saver...")
			app.PeriodicSaver.Stop()
			// Give periodic saver a moment to finish current operation
			time.Sleep(100 * time.Millisecond)
			loggerInstance.Debug("Periodic saver stopped successfully")
		}
		if err := filestorage.SaveOnShutdown(ctx, app.Repository, app.FileStorage); err != nil {
			loggerInstance.Error("Failed to save metrics on shutdown", zap.Error(err))
		}
	}

	loggerInstance.Debug("Shutting down server...")
	if err := app.Server.Shutdown(ctx); err != nil {
		loggerInstance.Error("Server shutdown failed", zap.Error(err))
		return err
	}

	if app.AuditPublisher != nil {
		if err := app.AuditPublisher.Close(ctx); err != nil {
			loggerInstance.Error("Audit shutdown failed", zap.Error(err))
		}
	}
	for _, o := range app.AuditObservers {
		if err := o.Close(); err != nil {
			loggerInstance.Error("Audit observer close failed", zap.Error(err))
		}
	}

	loggerInstance.Debug("Server stopped gracefully")
	return nil
}
