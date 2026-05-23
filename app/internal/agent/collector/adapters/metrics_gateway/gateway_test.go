package metrics_gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/compression"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/integrity"

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
