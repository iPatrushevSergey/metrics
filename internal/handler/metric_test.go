package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/internal/audit"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func floatp(f float64) *float64 { return &f }
func intp(i int64) *int64       { return &i }

func TestMetricHandlerUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
			name:       "invalid metric type",
			requestURL: "/update/unknownType/testGauge/123.4",
			method:     http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "invalid metric type",
			},
		},
		{
			name:       "invalid metric value",
			requestURL: "/update/gauge/testGauge/test",
			method:     http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "invalid metric value",
			},
		},
		{
			name:       "missing metric name",
			requestURL: "/update/counter/%20%20/10",
			method:     http.MethodPost,
			want: want{
				statusCode: http.StatusNotFound,
				body:       "The metric name is missing",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := inmemory.NewMemStorageMetricRepository()
			metricService := service.NewMetricService(repo)
			testLogger := logger.NewZapLoggerAdapter(zap.NewNop())
			metricHandler := NewMetricHandler(metricService, testLogger, audit.NewPublisher(testLogger))

			router := gin.New()
			router.POST("/update/:type/:name/:value", metricHandler.Update)

			request := httptest.NewRequest(tt.method, tt.requestURL, nil)
			request.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

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

func TestMetricHandlerGetValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	initialState := map[string]model.Metric{
		"Gauge":   {ID: "g1", MType: model.Gauge, Value: floatp(101.1)},
		"Counter": {ID: "c1", MType: model.Counter, Delta: intp(11)},
		"nil":     {ID: "g2", MType: model.Gauge, Value: nil},
	}

	type want struct {
		statusCode int
		body       string
	}

	tests := []struct {
		name       string
		requestURL string
		want       want
	}{
		{
			name:       "success get gauge",
			requestURL: "/value/gauge/Gauge",
			want:       want{statusCode: http.StatusOK, body: "101.1"},
		},
		{
			name:       "success get counter",
			requestURL: "/value/counter/Counter",
			want:       want{statusCode: http.StatusOK, body: "11"},
		},
		{
			name:       "no found metric",
			requestURL: "/value/gauge/nonexistent",
			want:       want{statusCode: http.StatusNotFound, body: "metric not found"},
		},
		{
			name:       "invalid metric type",
			requestURL: "/value/nonexistent/Counter",
			want:       want{statusCode: http.StatusBadRequest, body: "invalid metric type"},
		},
		{
			name:       "internal error on nil value",
			requestURL: "/value/gauge/nil",
			want:       want{statusCode: http.StatusInternalServerError, body: "internal service error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := inmemory.NewMemStorageMetricRepository()
			typedRepo := repo.(*inmemory.MemStorageMetricRepository)
			typedRepo.DB = initialState

			metricService := service.NewMetricService(typedRepo)
			testLogger := logger.NewZapLoggerAdapter(zap.NewNop())
			metricHandler := NewMetricHandler(metricService, testLogger, audit.NewPublisher(testLogger))

			router := gin.New()
			router.GET("/value/:type/:name", metricHandler.GetValue)

			request := httptest.NewRequest(http.MethodGet, tt.requestURL, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			bodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, string(bodyBytes))
		})
	}
}

func TestMetricHandlerGetJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	initialState := map[string]model.Metric{
		"gauge":   {ID: "gauge", MType: model.Gauge, Value: floatp(101.1)},
		"counter": {ID: "counter", MType: model.Counter, Delta: intp(11)},
		"nil":     {ID: "nil", MType: model.Gauge, Value: nil},
	}

	type want struct {
		statusCode int
		body       string
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "success get gauge",
			body: `{"id":"gauge", "type":"gauge"}`,
			want: want{statusCode: http.StatusOK, body: `{"id":"gauge","type":"gauge","value":101.1}`},
		},
		{
			name: "success get counter",
			body: `{"id":"counter", "type":"counter"}`,
			want: want{statusCode: http.StatusOK, body: `{"id":"counter","type":"counter","delta":11}`},
		},
		{
			name: "no found metric",
			body: `{"id":"notfound", "type":"counter"}`,
			want: want{statusCode: http.StatusNotFound, body: `{"error":"metric not found"}`},
		},
		{
			name: "invalid metric type",
			body: `{"id":"counter", "type":"notfound"}`,
			want: want{statusCode: http.StatusBadRequest, body: `{"error":"invalid metric type"}`},
		},
		{
			name: "internal error on nil value",
			body: `{"id":"nil", "type":"gauge"}`,
			want: want{statusCode: http.StatusInternalServerError, body: `{"error":"internal service error"}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := inmemory.NewMemStorageMetricRepository()
			typedRepo := repo.(*inmemory.MemStorageMetricRepository)
			typedRepo.DB = initialState

			metricService := service.NewMetricService(typedRepo)
			testLogger := logger.NewZapLoggerAdapter(zap.NewNop())
			metricHandler := NewMetricHandler(metricService, testLogger, audit.NewPublisher(testLogger))

			router := gin.New()
			router.POST("/value", metricHandler.GetJSON)

			reqBody := bytes.NewReader([]byte(tt.body))

			request := httptest.NewRequest(http.MethodPost, "/value", reqBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			bodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, string(bodyBytes))
		})
	}
}

func TestMetricHandlerGetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	initialState := map[string]model.Metric{
		"gauge1":  {ID: "g1", MType: model.Gauge, Value: floatp(99.9)},
		"counter": {ID: "c1", MType: model.Counter, Delta: intp(10)},
		"gauge2":  {ID: "g2", MType: model.Gauge, Value: floatp(50.5)},
		"nil":     {ID: "g3", MType: model.Gauge, Value: nil},
	}

	type want struct {
		statusCode     int
		contentType    string
		bodyContain    []string
		bodyNotContain []string
	}

	tests := []struct {
		name      string
		repoState map[string]model.Metric
		want      want
	}{
		{
			name:      "success get all metrics",
			repoState: initialState,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/html; charset=utf-8",
				bodyContain: []string{
					"<h1>All metrics</h1>",
					"<li><b>counter:</b> 10</li>",
					"<li><b>gauge1:</b> 99.9</li>",
					"<li><b>gauge2:</b> 50.5</li>",
				},
				bodyNotContain: []string{
					"nil",
				},
			},
		},
		{
			name:      "success get empty list",
			repoState: map[string]model.Metric{},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/html; charset=utf-8",
				bodyContain: []string{
					"<h1>All metrics</h1>",
					"<ul>",
					"</ul>",
				},
				bodyNotContain: []string{
					"<li>",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			repo := inmemory.NewMemStorageMetricRepository()
			typedRepo := repo.(*inmemory.MemStorageMetricRepository)
			typedRepo.DB = tt.repoState

			metricService := service.NewMetricService(typedRepo)
			testLogger := logger.NewZapLoggerAdapter(zap.NewNop())
			metricHandler := NewMetricHandler(metricService, testLogger, audit.NewPublisher(testLogger))

			router := gin.New()
			router.GET("/", metricHandler.GetAll)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			bodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			body := string(bodyBytes)

			for _, str := range tt.want.bodyContain {
				assert.Contains(t, body, str)
			}

			for _, str := range tt.want.bodyNotContain {
				assert.NotContains(t, body, str)
			}
		})
	}
}
