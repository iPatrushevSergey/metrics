package component_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	serverdto "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/http/dto"
	"github.com/iPatrushevSergey/metrics/app/tests/testsupport"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Server metrics component: HTTP → use cases → inmemory (no agent, no Postgres).
func TestServerMetricsComponent_batchUpsertAndGetJSON(t *testing.T) {
	srv := testsupport.StartMetricsServer(t)
	client := srv.Client()

	batch, err := json.Marshal([]serverdto.MetricRequest{
		{ID: "HeapAlloc", MType: "gauge", Value: ptr(512.0)},
	})
	require.NoError(t, err)

	resp, err := client.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(batch))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	getBody, err := json.Marshal(serverdto.MetricRequest{ID: "HeapAlloc", MType: "gauge"})
	require.NoError(t, err)

	resp, err = client.Post(srv.URL+"/value/", "application/json", bytes.NewReader(getBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got serverdto.MetricResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, "HeapAlloc", got.ID)
	require.NotNil(t, got.Value)
	assert.InDelta(t, 512, *got.Value, 0.001)
}

func ptr(v float64) *float64 { return &v }
