package e2e_test

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

func TestMetricsAPI_updateAndReadGauge(t *testing.T) {
	baseURL := testsupport.StartMetricsServer(t).URL

	one := 1.0
	batch, err := json.Marshal([]serverdto.MetricRequest{
		{ID: "PollCount", MType: "counter", Delta: ptr(int64(3))},
		{ID: "RandomValue", MType: "gauge", Value: &one},
	})
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/updates/", "application/json", bytes.NewReader(batch))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	getBody, err := json.Marshal(serverdto.MetricRequest{ID: "RandomValue", MType: "gauge"})
	require.NoError(t, err)

	resp, err = http.Post(baseURL+"/value/", "application/json", bytes.NewReader(getBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got serverdto.MetricResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, "RandomValue", got.ID)
	require.NotNil(t, got.Value)
}

func ptr(v int64) *int64 { return &v }
