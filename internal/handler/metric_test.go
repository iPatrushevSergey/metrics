package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricHandlerUpdate(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}

	tests := []struct {
		name       string
		requestURL string
		method     string
		want       want
	}{
		{
			name:       "success update gauge",
			requestURL: "/update/gauge/testGauge/123.4",
			method:     http.MethodPost,
			want:       want{statusCode: http.StatusOK},
		},
		{
			name:       "success update counter",
			requestURL: "/update/counter/testCounter/10",
			method:     http.MethodPost,
			want:       want{statusCode: http.StatusOK},
		},
		{
			name:       "invalid method",
			requestURL: "/update/gauge/testGauge/123.4",
			method:     http.MethodGet,
			want:       want{statusCode: http.StatusMethodNotAllowed},
		},
		{
			name:       "invalid metric type",
			requestURL: "/update/unknownType/testGauge/123.4",
			method:     http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "Invalid metric type\n",
			},
		},
		{
			name:       "invalid gauge value",
			requestURL: "/update/gauge/testGauge/test",
			method:     http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "Invalid gauge value\n",
			},
		},
		{
			name:       "invalid counter value",
			requestURL: "/update/counter/testCounter/test",
			method:     http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "Invalid counter value\n",
			},
		},
		{
			name:       "missing metric name",
			requestURL: "/update/counter/%20%20/10",
			method:     http.MethodPost,
			want: want{
				statusCode: http.StatusNotFound,
				body:       "The metric name is missing\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := inmemory.NewMemStorageMetricRepository()
			metricService := service.NewMetricService(repo)
			metricHandler := NewMetricHandler(metricService)

			mux := http.NewServeMux()
			mux.HandleFunc("/update/{type}/{name}/{value}", metricHandler.Update)

			request := httptest.NewRequest(tt.method, tt.requestURL, nil)
			request.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.body != "" {
				bodyBates, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.body, string(bodyBates))
			}
		})
	}
}
