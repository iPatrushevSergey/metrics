package bootstrap

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/adapters/metrics_gateway"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/adapters/sampler"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/presentation/worker"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/config"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/compression"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/http_client"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/integrity"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/retry"
)

// AgentApp holds the configuration and logger.
type AgentApp struct {
	cfg config.Config
	log port.Logger
}

// NewAgentApp initializes a new AgentApp.
func NewAgentApp(cfg config.Config, log port.Logger) *AgentApp {
	return &AgentApp{cfg: cfg, log: log}
}

// Run runs the agent application.
func (a *AgentApp) Run(shutdownCtx context.Context) error {
	pollCtx, cancelPoll := context.WithCancel(context.Background())
	sendCtx, cancelSend := context.WithCancel(context.Background())
	defer cancelSend()
	defer cancelPoll()

	go func() {
		<-shutdownCtx.Done()
		a.log.Info("shutdown signal received, stopping poll loops...")
		cancelPoll()
	}()

	var encryptor encryption.RSAEncryptor
	if pemPath := strings.TrimSpace(a.cfg.Agent.CryptoKey); pemPath != "" {
		pub, err := encryption.LoadRSAPublicKeyFromFile(pemPath)
		if err != nil {
			return fmt.Errorf("RSA public key: %w", err)
		}
		encryptor = encryption.NewRSAEncryptorWithPublic(pub)
	} else {
		a.log.Warn("RSA body encryption disabled (empty agent crypto key)")
		encryptor = encryption.NewRSAEncryptorWithPublic(nil)
	}

	metricsRepo := inmemory.NewMetricsRepository()
	metricsSampler := sampler.NewMetricsSampler()
	metricsGateway := metrics_gateway.NewGateway(
		a.cfg.Agent.MetricsGatewayConfig,
		&http.Client{Timeout: a.cfg.Agent.MetricsGatewayConfig.HTTPTimeout},
		compression.NewGzipCompressor(),
		encryptor,
		integrity.NewSHA256Hasher(a.cfg.Agent.Key),
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

	var totalMetricsGateway port.MetricsGateway = metricsGateway
	var bufferedMetricsGateway *metrics_gateway.BufferedMetricsGateway
	if a.cfg.Agent.RateLimit > 0 {
		var err error
		bufferedMetricsGateway, err = metrics_gateway.NewBufferedMetricsGateway(metricsGateway, a.log, a.cfg.Agent.RateLimit, sendCtx)
		if err != nil {
			return fmt.Errorf("buffered metrics gateway: %w", err)
		}
		bufferedMetricsGateway.Start()
		totalMetricsGateway = bufferedMetricsGateway
	}

	ucf := NewUseCaseFactory(
		WithMetricsRepo(metricsRepo),
		WithMetricsSampler(metricsSampler),
		WithMetricsGateway(totalMetricsGateway),
		WithLogger(a.log),
		WithRandFloat(rand.Float64),
		WithReportInterval(a.cfg.Agent.ReportInterval),
	)

	go worker.NewPollRuntimeWorker(ucf, a.log, a.cfg.Agent.PollInterval).Run(pollCtx)
	go worker.NewPollGopsutilWorker(ucf, a.log, a.cfg.Agent.PollInterval).Run(pollCtx)

	report := worker.NewReportWorker(sendCtx, ucf, a.log, a.cfg.Agent.ReportInterval)
	report.Run(pollCtx)
	report.Wait()

	if bufferedMetricsGateway != nil {
		bufferedMetricsGateway.Stop()
	}

	cancelSend()

	return shutdownCtx.Err()
}
