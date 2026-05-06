package bootstrap

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/adapters/metrics_client"
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
	metricsClient := metrics_client.NewClient(
		a.cfg.Agent.MetricsClientConfig,
		&http.Client{Timeout: a.cfg.Agent.MetricsClientConfig.HTTPTimeout},
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

	var totalMetricsClient port.MetricsClient = metricsClient
	var bufferedMetricsClient *metrics_client.BufferedClient
	if a.cfg.Agent.RateLimit > 0 {
		var err error
		bufferedMetricsClient, err = metrics_client.NewBufferedClient(metricsClient, a.log, a.cfg.Agent.RateLimit, sendCtx)
		if err != nil {
			return fmt.Errorf("buffered metrics client: %w", err)
		}
		bufferedMetricsClient.Start()
		totalMetricsClient = bufferedMetricsClient
	}

	ucf := NewUseCaseFactory(
		WithMetricsRepo(metricsRepo),
		WithMetricsSampler(metricsSampler),
		WithMetricsClient(totalMetricsClient),
		WithLogger(a.log),
		WithRandFloat(rand.Float64),
		WithReportInterval(a.cfg.Agent.ReportInterval),
	)

	go worker.RunPoolTickerLoop(pollCtx, ucf.PollRuntimeTick(), "poll_runtime", a.log, a.cfg.Agent.PollInterval)
	go worker.RunPoolTickerLoop(pollCtx, ucf.PollGopsutilTick(), "poll_gopsutil", a.log, a.cfg.Agent.PollInterval)

	report := worker.NewReportLoop(sendCtx, ucf.ReportBatchTick(), a.log, a.cfg.Agent.ReportInterval)
	report.RunReportTickerLoop(pollCtx)

	report.WaitSendsWg()

	if bufferedMetricsClient != nil {
		bufferedMetricsClient.Stop()
	}

	cancelSend()

	return shutdownCtx.Err()
}
