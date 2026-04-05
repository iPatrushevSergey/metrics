package bootstrap

import (
	"crypto/rsa"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/iPatrushevSergey/metrics/internal/audit"
	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/reqcrypto"
	"github.com/iPatrushevSergey/metrics/internal/service"
)

// InitializeApp initializes all application components and returns configured App
func InitializeApp(cfg config.ServerConfig, loggerAdapter logger.Logger) (*App, error) {
	loggerAdapter.Debug("starting server with config", zap.Object("cfg details", &cfg))

	// Initialize repository
	repoConfig, err := InitializeRepository(cfg, loggerAdapter)
	if err != nil {
		return nil, err
	}

	// Create service and handler
	metricService := service.NewMetricService(repoConfig.Repository)

	var observers []audit.Observer

	if cfg.AuditFilePath != "" {
		fo, err := audit.NewFileObserver(cfg.AuditFilePath)
		if err != nil {
			return nil, err
		}
		observers = append(observers, fo)
	}
	if cfg.AuditURL != "" {
		ho, err := audit.NewHTTPObserver(cfg.AuditURL, &http.Client{
			Timeout: cfg.AuditHTTPTimeout,
		})
		if err != nil {
			return nil, err
		}
		observers = append(observers, ho)
	}

	auditPublisher := audit.NewPublisher(loggerAdapter, observers...)

	metricHandler := handler.NewMetricHandler(metricService, loggerAdapter, auditPublisher)

	var priv *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		var err error
		priv, err = reqcrypto.LoadRSAPrivateKeyFromFile(cfg.CryptoKey)
		if err != nil {
			return nil, fmt.Errorf("load RSA private key: %w", err)
		}
	}

	// Setup router
	router := SetupRouter(metricHandler, cfg, loggerAdapter, priv)

	// Create HTTP server
	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	return &App{
		Server:         server,
		DB:             repoConfig.DB,
		Repository:     repoConfig.Repository,
		FileStorage:    repoConfig.FileStorage,
		PeriodicSaver:  repoConfig.PeriodicSaver,
		AuditPublisher: auditPublisher,
		AuditObservers: observers,
	}, nil
}

// Run initializes and runs the application
func Run(cfg config.ServerConfig) error {
	initializedLogger, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		return err
	}
	defer initializedLogger.Sync()

	loggerAdapter := logger.NewZapLoggerAdapter(initializedLogger)

	// Initialize application with logger adapter
	app, err := InitializeApp(cfg, loggerAdapter)
	if err != nil {
		return err
	}

	if app.DB != nil {
		defer app.DB.Close()
	}

	serverErrCh := StartServer(app.Server, loggerAdapter)
	return WaitForShutdown(app, cfg, loggerAdapter, serverErrCh)
}
