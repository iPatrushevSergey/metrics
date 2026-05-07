package metrics_client

import (
	"context"
	"net/http"

	clientport "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/adapters/metrics_client/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/compression"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/integrity"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/adapters/retry"
)

var (
	_ port.MetricsClient    = (*Client)(nil)
	_ clientport.Compressor = compression.GzipCompressor{}
	_ clientport.Encryptor  = encryption.RSAEncryptor{}
	_ clientport.Hasher     = integrity.SHA256Hasher{}
)

// preparedRequest contains the URL, body bytes, and headers for sending the request.
type preparedRequest struct {
	url     string
	body    []byte
	headers http.Header
}

// Client is the net/http implementation for the metrics server.
type Client struct {
	httpClient *http.Client
	baseURL    string
	compressor clientport.Compressor
	encryptor  clientport.Encryptor
	hasher     clientport.Hasher
	retryOpts  []retry.RetryOption
}

// NewClient initializes a new metrics client.
func NewClient(
	cfg MetricsClientConfig,
	httpClient *http.Client,
	compressor clientport.Compressor,
	encryptor clientport.Encryptor,
	hasher clientport.Hasher,
	retryOpts ...retry.RetryOption,
) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    cfg.Address,
		compressor: compressor,
		encryptor:  encryptor,
		hasher:     hasher,
		retryOpts:  retryOpts,
	}
}

// MetricsUpdateBatch sends a batch of metric updates to the metrics server.
func (c *Client) MetricsUpdateBatch(ctx context.Context, metrics []dto.MetricUpdateInput) error {
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
