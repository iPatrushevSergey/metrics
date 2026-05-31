package contract_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/adapters/metrics_gateway"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/compression"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/integrity"
	serverdto "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/http/dto"
	"github.com/iPatrushevSergey/metrics/app/tests/testsupport"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Consumer contract: agent must send a batch the server accepts.
func TestAgentMetricsGateway_ServerUpdatesContract(t *testing.T) {
	srv := testsupport.StartMetricsServer(t)

	gw := metricsgateway.NewGateway(
		metricsgateway.MetricsGatewayConfig{Address: srv.URL},
		srv.Client(),
		compression.GzipCompressor{},
		encryption.RSAEncryptor{},
		integrity.SHA256Hasher{},
	)

	val := 42.0
	err := gw.MetricsUpdateBatch(context.Background(), []dto.MetricUpdateInput{
		{ID: "Alloc", MType: "gauge", Value: &val},
	})
	require.NoError(t, err)

	body, err := json.Marshal(serverdto.MetricRequest{ID: "Alloc", MType: "gauge"})
	require.NoError(t, err)

	resp, err := srv.Client().Post(srv.URL+"/value/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got serverdto.MetricResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, "Alloc", got.ID)
	assert.Equal(t, "gauge", got.MType)
	require.NotNil(t, got.Value)
	assert.InDelta(t, 42, *got.Value, 0.001)
}

// Producer contract: server must accept gzip-compressed JSON batch.
func TestServerUpdates_acceptsGzipJSONBatch(t *testing.T) {
	gz := compression.NewGzipCompressor()
	payload, err := json.Marshal([]serverdto.MetricRequest{
		{ID: "a", MType: "gauge", Value: ptr(1.0)},
	})
	require.NoError(t, err)

	var compressed bytes.Buffer
	zw := gz.CompressWriter(&compressed)
	_, err = zw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/updates", r.URL.Path)
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		plain, err := gz.DecompressReader(bytes.NewReader(raw))
		require.NoError(t, err)
		body, err := io.ReadAll(plain)
		require.NoError(t, err)
		plain.Close()

		var batch []serverdto.MetricRequest
		require.NoError(t, json.Unmarshal(body, &batch))
		require.Len(t, batch, 1)
		assert.Equal(t, "a", batch[0].ID)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/updates", bytes.NewReader(compressed.Bytes()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func ptr(v float64) *float64 { return &v }
