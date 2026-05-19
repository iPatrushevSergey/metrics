package bootstrap

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/logger"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/retry"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/option"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/config"
	metricfile "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/adapters/repository/file"
	metricinmemory "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/adapters/repository/inmemory"
	metricpostgres "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/adapters/repository/postgres"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/service"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/presentation/lifecycle"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StorageMode defines the metrics persistence mode.
type StorageMode int

const (
	StoragePostgres StorageMode = iota // PostgreSQL database
	StorageFile                        // File storage
	StorageMemory                      // In-memory storage
)

// Run starts the server.
func Run() error {
	// Load config.
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Define storage mode.
	var storageMode StorageMode
	switch {
	case strings.TrimSpace(cfg.DB.Pool.URI) != "":
		storageMode = StoragePostgres
	case strings.TrimSpace(cfg.Server.FileStoragePath) != "":
		storageMode = StorageFile
	default:
		storageMode = StorageMemory
	}

	// Initialize logger.
	var _ port.Logger = (*logger.ZapLogger)(nil)
	zl, err := logger.NewZapLogger(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer zl.Sync()

	// Log server startup details.
	zl.Debug("starting server",
		"address", cfg.Server.Address,
		"storage_mode", storageMode,
		"database_configured", cfg.DB.Pool.URI != "",
		"db_max_conns", cfg.DB.Pool.MaxConns,
		"db_min_conns", cfg.DB.Pool.MinConns,
		"db_retry_max_retries", cfg.DB.Retry.MaxRetries,
		"enable_retry", cfg.Server.EnableRetry,
		"log_level", cfg.Logger.Level,
		"shutdown_timeout", cfg.Server.ShutdownTimeout,
		"store_interval", cfg.Server.StoreInterval,
		"store_file", cfg.Server.FileStoragePath,
		"restore", cfg.Server.Restore,
		"key_configured", cfg.Server.Key != "",
		"crypto_key_configured", cfg.Server.CryptoKey != "",
		"audit_file_configured", cfg.Server.AuditFilePath != "",
		"audit_url_configured", cfg.Server.AuditURL != "",
		"audit_http_timeout", cfg.Server.AuditHTTPTimeout,
	)

	// Initialize database pool.
	var pool *pgxpool.Pool
	if storageMode == StoragePostgres {
		pool, err = postgres.NewPool(context.Background(), cfg.DB.Pool)
		if err != nil {
			return fmt.Errorf("init database pool: %w", err)
		}
		defer pool.Close()
	}

	// Initialize transactor.
	var transactor port.Transactor
	switch storageMode {
	case StoragePostgres:
		maxRetries := cfg.DB.Retry.MaxRetries
		if !cfg.Server.EnableRetry {
			maxRetries = 0
		}
		retryOpts := []retry.RetryOption{retry.WithMaxRetries(maxRetries)}
		if maxRetries > 0 {
			retryOpts = append(retryOpts, retry.WithBackoffFunc(func(attempt int) time.Duration {
				switch attempt {
				case 0:
					return 1 * time.Second
				case 1:
					return 3 * time.Second
				default:
					return 5 * time.Second
				}
			}))
		}
		transactor = postgres.NewTransactor(pool, retryOpts...)
	case StorageFile, StorageMemory:
		transactor = inmemory.NewTransactor()
	}

	// Load RSA private key.
	var priv *rsa.PrivateKey
	if pemPath := strings.TrimSpace(cfg.Server.CryptoKey); pemPath != "" {
		priv, err = encryption.LoadRSAPrivateKeyFromFile(pemPath)
		if err != nil {
			return fmt.Errorf("load RSA private key: %w", err)
		}
	}

	// Initialize metric repository.
	var metricRepo port.MetricRepository
	switch storageMode {
	case StoragePostgres:
		pgTr, ok := transactor.(*postgres.Transactor)
		if !ok {
			return fmt.Errorf("postgres storage requires *postgres.Transactor, got %T", transactor)
		}
		metricRepo = metricpostgres.NewMetricPostgresRepository(pgTr)
	default:
		metricRepo = metricinmemory.NewMetricMemoryRepository()
	}

	// Build options for the use case factory.
	factoryOpts := []option.Option[factoryParams]{
		WithMetricRepo(metricRepo),
		WithMetricSvc(service.MetricService{}),
		WithTransactor(transactor),
	}

	// Initialize metric file repository with options.
	switch storageMode {
	case StorageFile:
		// File repo: restore on startup; periodic or sync snapshots while running; snapshot on shutdown.
		metricFileRepo := metricfile.NewMetricFileRepository(cfg.Server.FileStoragePath)
		factoryOpts = append(factoryOpts,
			WithMetricFileRepo(metricFileRepo),
			WithSyncFileWrites(cfg.Server.StoreInterval == 0), // interval==0: persist after each mutation
		)
	case StoragePostgres, StorageMemory:
		// FILE_STORAGE_PATH, STORE_INTERVAL and RESTORE are not used.
	}

	// Initialize use case factory.
	useCases := NewUseCaseFactory(factoryOpts...)

	// Initialize router.
	router, err := NewRouter(useCases, zl, cfg.Server.Key, priv)
	if err != nil {
		return fmt.Errorf("router: %w", err)
	}

	app := &lifecycle.App{
		Server: &http.Server{
			Addr:    cfg.Server.Address,
			Handler: router,
		},
		UseCases:        useCases,
		Log:             zl,
		ShutdownTimeout: cfg.Server.ShutdownTimeout,
		FileStorage:     storageMode == StorageFile,
		Restore:         cfg.Server.Restore,
		StoreInterval:   cfg.Server.StoreInterval,
	}

	app.Start()
	return app.Stop()
}
