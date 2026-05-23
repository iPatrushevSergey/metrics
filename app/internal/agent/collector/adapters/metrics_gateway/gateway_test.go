package metrics_gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/compression"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/http_client"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/integrity"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/retry"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsGateway_MetricsUpdateBatch_empty(t *testing.T) {
	gw := NewGateway(
		MetricsGatewayConfig{Address: "http://127.0.0.1:8080"},
		http.DefaultClient,
		compression.GzipCompressor{},
		encryption.RSAEncryptor{},
		integrity.SHA256Hasher{},
	)

	err := gw.MetricsUpdateBatch(context.Background(), nil)
	assert.NoError(t, err)
}

func TestMetricsGateway_MetricsUpdateBatch_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	gw := NewGateway(
		MetricsGatewayConfig{Address: srv.URL},
		srv.Client(),
		compression.GzipCompressor{},
		encryption.RSAEncryptor{},
		integrity.SHA256Hasher{},
	)

	val := 1.5
	err := gw.MetricsUpdateBatch(context.Background(), []dto.MetricUpdateInput{
		{ID: "Alloc", MType: "gauge", Value: &val},
	})
	require.NoError(t, err)
}

func TestMetricsGateway_MetricsUpdateBatch_serverError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("fail"))
	}))
	t.Cleanup(srv.Close)

	gw := NewGateway(
		MetricsGatewayConfig{Address: srv.URL},
		srv.Client(),
		compression.GzipCompressor{},
		encryption.RSAEncryptor{},
		integrity.SHA256Hasher{},
	)

	val := 1.0
	err := gw.MetricsUpdateBatch(context.Background(), []dto.MetricUpdateInput{
		{ID: "a", MType: "gauge", Value: &val},
	})
	assert.Error(t, err)
}

func TestMetricsGateway_MetricsUpdateBatch_retriesOn503(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	gw := NewGateway(
		MetricsGatewayConfig{Address: srv.URL},
		srv.Client(),
		compression.GzipCompressor{},
		encryption.RSAEncryptor{},
		integrity.SHA256Hasher{},
		retry.WithMaxRetries(2),
		retry.WithConstantBackoff(0),
		retry.WithRetriableCheck(http_client.IsRetriable),
	)

	val := 1.0
	err := gw.MetricsUpdateBatch(context.Background(), []dto.MetricUpdateInput{
		{ID: "a", MType: "gauge", Value: &val},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
}
