// Package bootstrap is the composition root for the metrics HTTP server.
package bootstrap

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/http_client"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/retry"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/migrate"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/option"
	"github.com/iPatrushevSergey/metrics/app/internal/server/config"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/audit"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/audit_gateway"
	fileaudit "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/audit"
	filemetrics "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/metrics"
	metricinmemory "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/inmemory"
	metricpostgres "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/postgres"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/lifecycle"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StorageMode defines the metrics persistence mode.
type StorageMode int

const (
	StoragePostgres StorageMode = iota // PostgreSQL database
	StorageFile                        // File storage
	StorageMemory                      // In-memory storage
)

// Run loads configuration, wires dependencies and runs the application.
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
		"audit_sub_size", cfg.Audit.AuditSubSize,
	)

	// Initialize database pool.
	var pool *pgxpool.Pool
	if storageMode == StoragePostgres {
		if err := migrate.Up(cfg.DB.Pool.URI, migrate.MigrationsMetricsDir()); err != nil {
			return fmt.Errorf("apply migrations: %w", err)
		}
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
		metricFileRepo := filemetrics.NewMetricFileRepository(cfg.Server.FileStoragePath)
		factoryOpts = append(factoryOpts,
			WithMetricFileRepo(metricFileRepo),
			WithSyncFileWrites(cfg.Server.StoreInterval == 0), // interval==0: persist after each mutation
		)
	case StoragePostgres, StorageMemory:
		// FILE_STORAGE_PATH, STORE_INTERVAL and RESTORE are not used.
	}

	// Initialize audit components.
	var (
		auditPublisher    port.AuditPublisher
		auditFileRepo     port.AuditFileRepository
		auditGateway      port.AuditGateway
		auditFileEvents   <-chan dto.AuditEvent
		auditRemoteEvents <-chan dto.AuditEvent
	)
	auditFilePath := strings.TrimSpace(cfg.Server.AuditFilePath)
	auditURL := strings.TrimSpace(cfg.Server.AuditURL)

	// Initialize audit publisher.
	if auditFilePath != "" || auditURL != "" {
		auditPublisher = audit.NewAuditEventPublisher(zl, cfg.Audit.AuditSubSize)

		// Initialize audit file repository.
		if auditFilePath != "" {
			auditFileRepo, err = fileaudit.NewAuditFileRepository(auditFilePath)
			if err != nil {
				return fmt.Errorf("audit file repository: %w", err)
			}
			factoryOpts = append(factoryOpts, WithAuditFileRepo(auditFileRepo))
			auditFileEvents, err = auditPublisher.Subscribe("file")
			if err != nil {
				return fmt.Errorf("subscribe audit file worker: %w", err)
			}
		}

		// Initialize audit remote gateway.
		if auditURL != "" {
			auditGateway, err = audit_gateway.NewAuditRemoteGateway(
				audit_gateway.AuditGatewayConfig{
					URL:         auditURL,
					HTTPTimeout: cfg.Server.AuditHTTPTimeout,
				},
				&http.Client{Timeout: cfg.Server.AuditHTTPTimeout},
				retry.WithRetriableCheck(http_client.IsRetriable),
				retry.WithMaxRetries(3),
				retry.WithBackoffFunc(func(attempt int) time.Duration {
					switch attempt {
					case 0:
						return 1 * time.Second
					case 1:
						return 3 * time.Second
					default:
						return 5 * time.Second
					}
				}),
			)
			if err != nil {
				return fmt.Errorf("audit remote gateway: %w", err)
			}
			factoryOpts = append(factoryOpts, WithAuditGateway(auditGateway))
			auditRemoteEvents, err = auditPublisher.Subscribe("remote")
			if err != nil {
				return fmt.Errorf("subscribe audit remote worker: %w", err)
			}
		}
		factoryOpts = append(factoryOpts, WithAuditPublisher(auditPublisher))
	}

	// Build use case factory.
	useCases := NewUseCaseFactory(factoryOpts...)

	// Initialize trusted subnet.
	var trustedSubnet *net.IPNet
	if cidr := cfg.Server.TrustedSubnet; cidr != "" {
		_, trustedSubnet, err = net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("trusted subnet: %w", err)
		}
	}

	// Initialize router.
	router, err := NewRouter(useCases, zl, cfg.Server.Key, priv, trustedSubnet)
	if err != nil {
		return fmt.Errorf("router: %w", err)
	}

	// Initialize application.
	app := &lifecycle.App{
		Server: &http.Server{
			Addr:    cfg.Server.Address,
			Handler: router,
		},
		UseCases:          useCases,
		Log:               zl,
		ShutdownTimeout:   cfg.Server.ShutdownTimeout,
		FileStorage:       storageMode == StorageFile,
		Restore:           cfg.Server.Restore,
		StoreInterval:     cfg.Server.StoreInterval,
		AuditPublisher:    auditPublisher,
		AuditFileRepo:     auditFileRepo,
		AuditFileEvents:   auditFileEvents,
		AuditRemoteEvents: auditRemoteEvents,
	}

	app.Start()
	return app.Stop()
}
