package bootstrap

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx driver for database/sql

	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/filestorage"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/repository"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/internal/repository/postgres"
)

const dbPingTimeout = 3 * time.Second

// RepositoryConfig holds configuration for repository initialization
type RepositoryConfig struct {
	DB            *sql.DB
	Repository    repository.MetricRepository
	FileStorage   *filestorage.FileStorage
	PeriodicSaver *filestorage.PeriodicSaver
}

// InitializeRepository creates and configures the repository based on configuration
func InitializeRepository(cfg config.ServerConfig, loggerInstance logger.Logger) (*RepositoryConfig, error) {
	var (
		repo          repository.MetricRepository
		fs            *filestorage.FileStorage
		periodicSaver *filestorage.PeriodicSaver
		db            *sql.DB
	)

	if cfg.DatabaseDSN != "" {
		// PostgreSQL
		loggerInstance.Debug("Using PostgreSQL storage")

		if err := postgres.RunMigrations(cfg.DatabaseDSN); err != nil {
			return nil, err
		}

		var err error
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), dbPingTimeout)
		defer cancel()

		if err = db.PingContext(ctx); err != nil {
			db.Close()
			return nil, err
		}

		// Creating a basic PostgreSQL repository
		baseRepo := postgres.NewPostgresMetricRepository(db)

		// Optionally wrap in retry decorator if enabled in config
		if cfg.EnableRetry {
			loggerInstance.Debug("Retry logic enabled for PostgreSQL operations")
			repo = postgres.NewRetryRepository(baseRepo, postgres.DefaultRetryConfig())
		} else {
			repo = baseRepo
		}
		fs = nil
	} else {
		// Inmemory
		loggerInstance.Debug("Using In-Memory storage with FileStorage")

		memRepo := inmemory.NewMemStorageMetricRepository()
		fs = filestorage.NewFileStorage(cfg.FileStoragePath)

		if cfg.Restore {
			loggerInstance.Debug("Restoring metrics from file", zap.String("path", cfg.FileStoragePath))
			restoredMetrics, err := fs.Load()
			if err != nil {
				loggerInstance.Error("Failed to restore metrics", zap.Error(err))
			} else {
				ctxRestore := context.Background()
				for _, m := range restoredMetrics {
					if err := memRepo.Create(ctxRestore, m); err != nil {
						loggerInstance.Warn("Failed to restore metric", zap.String("id", m.ID), zap.Error(err))
					}
				}
				loggerInstance.Debug("Metrics restored successfully", zap.Int("count", len(restoredMetrics)))
			}
		}

		// Defining the inmemory type (sync or async)
		if cfg.StoreInterval == 0 {
			loggerInstance.Debug("Sync saving mode enable (StorageInterval=0)")
			repo = repository.NewSyncFileRepository(memRepo, fs)
		} else {
			repo = memRepo
		}

		if cfg.StoreInterval > 0 {
			periodicSaver = filestorage.NewPeriodicSaver(repo, fs, cfg.StoreInterval)
			periodicSaver.Start()
		}
	}

	return &RepositoryConfig{
		DB:            db,
		Repository:    repo,
		FileStorage:   fs,
		PeriodicSaver: periodicSaver,
	}, nil
}
