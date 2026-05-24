// Package bootstrap is the composition root for the metrics collection agent.
package bootstrap

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/adapters/metrics_gateway"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/adapters/metrics_grpc"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/adapters/sampler"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/presentation/lifecycle"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/config"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/compression"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/http_client"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/integrity"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/retry"
)

// Run loads configuration, wires dependencies and runs the application.
func Run() error {
	// Load config.
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize logger.
	var _ port.Logger = (*logger.ZapLogger)(nil)
	log, err := logger.NewZapLogger(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer log.Sync()

	// Initialize worker contexts.
	pollCtx, cancelPoll := context.WithCancel(context.Background())
	sendCtx, cancelSend := context.WithCancel(context.Background())
	defer cancelPoll()
	defer cancelSend()

	// Load RSA public key.
	var encryptor encryption.RSAEncryptor
	if pemPath := strings.TrimSpace(cfg.Agent.CryptoKey); pemPath != "" {
		pub, err := encryption.LoadRSAPublicKeyFromFile(pemPath)
		if err != nil {
			return fmt.Errorf("RSA public key: %w", err)
		}
		encryptor = encryption.NewRSAEncryptorWithPublic(pub)
	} else {
		log.Warn("RSA body encryption disabled (empty agent crypto key)")
		encryptor = encryption.NewRSAEncryptorWithPublic(nil)
	}

	// Initialize metric repository.
	metricsRepo := inmemory.NewMetricsRepository()

	// Initialize metrics sampler.
	metricsSampler := sampler.NewMetricsSampler()

	// Initialize metrics gateway.
	var totalMetricsGateway port.MetricsGateway
	var gatewayCloser interface{ Close() error }

	switch cfg.Agent.ReportProtocol {
	case config.ReportProtocolGRPC:
		grpcGateway, err := metrics_grpc.NewGateway(cfg.Agent.MetricsGRPCGatewayConfig)
		if err != nil {
			return fmt.Errorf("metrics grpc gateway: %w", err)
		}
		totalMetricsGateway = grpcGateway
		gatewayCloser = grpcGateway
		log.Info("report protocol: grpc", "address", cfg.Agent.MetricsGRPCGatewayConfig.Address)
	default:
		totalMetricsGateway = metricsgateway.NewGateway(
			cfg.Agent.MetricsGatewayConfig,
			&http.Client{Timeout: cfg.Agent.MetricsGatewayConfig.HTTPTimeout},
			compression.NewGzipCompressor(),
			encryptor,
			integrity.NewSHA256Hasher(cfg.Agent.Key),
			retry.WithRetriableCheck(httpclient.IsRetriable),
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
		log.Info("report protocol: http", "address", cfg.Agent.MetricsGatewayConfig.Address)
	}

	// Initialize rate-limited metrics gateway.
	var bufferedMetricsGateway *metricsgateway.BufferedMetricsGateway
	if cfg.Agent.RateLimit > 0 {
		bufferedMetricsGateway, err = metricsgateway.NewBufferedMetricsGateway(
			totalMetricsGateway, log, cfg.Agent.RateLimit, sendCtx,
		)
		if err != nil {
			return fmt.Errorf("buffered metrics gateway: %w", err)
		}
		totalMetricsGateway = bufferedMetricsGateway
	}

	// Build use case factory.
	useCases := NewUseCaseFactory(
		WithMetricsRepo(metricsRepo),
		WithMetricsSampler(metricsSampler),
		WithMetricsGateway(totalMetricsGateway),
		WithLogger(log),
		WithRandFloat(rand.Float64),
		WithReportInterval(cfg.Agent.ReportInterval),
	)

	// Initialize application.
	app := &lifecycle.App{
		UseCases:        useCases,
		Log:             log,
		PollInterval:    cfg.Agent.PollInterval,
		ReportInterval:  cfg.Agent.ReportInterval,
		PollCtx:         pollCtx,
		SendCtx:         sendCtx,
		CancelPoll:      cancelPoll,
		CancelSend:      cancelSend,
		BufferedGateway: bufferedMetricsGateway,
		GatewayCloser:   gatewayCloser,
	}

	app.Start()
	return app.Stop()
}
