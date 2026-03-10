package filestorage

import (
	"context"
	"time"

	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/repository"
	"go.uber.org/zap"
)

// PeriodicSaver saves repository metrics to file at a fixed interval.
type PeriodicSaver struct {
	repo     repository.MetricRepository
	fs       *FileStorage
	interval time.Duration
	stopCh   chan struct{}
}

// NewPeriodicSaver returns a PeriodicSaver. Call Start to begin saving; call Stop before shutdown.
func NewPeriodicSaver(repo repository.MetricRepository, fs *FileStorage, interval time.Duration) *PeriodicSaver {
	return &PeriodicSaver{
		repo:     repo,
		fs:       fs,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the background goroutine that saves metrics at each interval.
func (ps *PeriodicSaver) Start() {
	go func() {
		ticker := time.NewTicker(ps.interval)
		defer ticker.Stop()

		logger.Log.Debug("Starting periodic metric saver", zap.Duration("interval", ps.interval))

		for {
			select {
			case <-ticker.C:
				ps.saveMetrics()
			case <-ps.stopCh:
				logger.Log.Debug("Periodic saver stopped")
				return
			}
		}
	}()
}

// Stop stops the periodic saver goroutine.
func (ps *PeriodicSaver) Stop() {
	close(ps.stopCh)
}

func (ps *PeriodicSaver) saveMetrics() {
	allMetrics, err := ps.repo.GetAll(context.Background())
	if err != nil {
		logger.Log.Error("Failed to get metrics for saving", zap.Error(err))
		return
	}

	if err := ps.fs.Save(allMetrics); err != nil {
		logger.Log.Error("Failed to save metrics to file", zap.Error(err))
	} else {
		logger.Log.Debug("Metrics saved to file")
	}
}

// SaveOnShutdown writes current repository metrics to file once (e.g. on graceful shutdown).
func SaveOnShutdown(ctx context.Context, repo repository.MetricRepository, fs *FileStorage) error {
	logger.Log.Debug("Saving metrics before shutdown...")
	allMetrics, err := repo.GetAll(ctx)
	if err != nil {
		logger.Log.Error("Getting metrics failed", zap.Error(err))
		return err
	}

	if err := fs.Save(allMetrics); err != nil {
		logger.Log.Error("Failed to save metrics on shutdown", zap.Error(err))
		return err
	}

	logger.Log.Debug("Metrics saved successfully on shutdown")
	return nil
}
