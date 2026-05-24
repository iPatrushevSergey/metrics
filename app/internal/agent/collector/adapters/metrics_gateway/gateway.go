// Package metrics_gateway sends metric batches to the metrics server.
package metrics_gateway

import (
	"context"
	"net/http"
	"strings"

	gatewayport "github.com/iPatrushevSergey/metrics/app/internal/agent/collector/adapters/metrics_gateway/port"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/compression"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/integrity"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/retry"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/netutil"
)

var (
	_ port.MetricsGateway    = (*metricsGateway)(nil)
	_ gatewayport.Compressor = compression.GzipCompressor{}
	_ gatewayport.Encryptor  = encryption.RSAEncryptor{}
	_ gatewayport.Hasher     = integrity.SHA256Hasher{}
)

// preparedRequest contains the URL, body bytes, and headers for sending the request.
type preparedRequest struct {
	url     string
	body    []byte
	headers http.Header
}

// metricsGateway is the net/http implementation for the metrics server.
type metricsGateway struct {
	httpClient *http.Client
	baseURL    string
	realIP     string
	compressor gatewayport.Compressor
	encryptor  gatewayport.Encryptor
	hasher     gatewayport.Hasher
	retryOpts  []retry.RetryOption
}

// NewGateway initializes the HTTP metrics gateway.
func NewGateway(
	cfg MetricsGatewayConfig,
	httpClient *http.Client,
	compressor gatewayport.Compressor,
	encryptor gatewayport.Encryptor,
	hasher gatewayport.Hasher,
	retryOpts ...retry.RetryOption,
) *metricsGateway {
	realIP := strings.TrimSpace(cfg.RealIP)
	if realIP == "" {
		if ip, err := netutil.HostIPv4(); err == nil {
			realIP = ip
		}
	}
	return &metricsGateway{
		httpClient: httpClient,
		baseURL:    cfg.Address,
		realIP:     realIP,
		compressor: compressor,
		encryptor:  encryptor,
		hasher:     hasher,
		retryOpts:  retryOpts,
	}
}

// MetricsUpdateBatch sends a batch of metric updates to the metrics server.
func (c *metricsGateway) MetricsUpdateBatch(ctx context.Context, metrics []dto.MetricUpdateInput) error {
	if len(metrics) == 0 {
		return nil
	}

	prepared, err := c.metricsUpdateBatchPrepare(metrics)
	if err != nil {
		return err
	}

	var resp *http.Response

	err = retry.DoWithRetry(ctx, func() error {
		r, sendErr := c.metricsUpdateBatchSend(ctx, prepared)
		if sendErr != nil {
			return sendErr
		}
		resp = r
		return nil
	}, c.retryOpts...)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return c.metricsUpdateBatchProcessResponse(resp)
}
