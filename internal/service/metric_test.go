package service

import (
	"testing"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func floatp(f float64) *float64 { return &f }
func intp(i int64) *int64       { return &i }

func TestMetricServiceGet(t *testing.T) {
	initialState := map[string]model.Metric{
		"gauge":   {ID: "g1", MType: model.Gauge, Value: floatp(10.5)},
		"counter": {ID: "c1", MType: model.Counter, Delta: intp(50)},
	}

	repo := inmemory.NewMemStorageMetricRepository()
	typedRepo := repo.(*inmemory.MemStorageMetricRepository)
	typedRepo.DB = initialState

	metricService := NewMetricService(typedRepo)

	t.Run("success get metric", func(t *testing.T) {
		metric, err := metricService.GetValue("gauge", "gauge")
		require.NoError(t, err)
		assert.Equal(t, "10.5", metric)
	})

	t.Run("error metric not found", func(t *testing.T) {
		_, err := metricService.GetValue("counter", "nonexistent")
		require.Error(t, err)
		assert.EqualError(t, err, "metric not found")
	})
}

func TestMetricServiceGetAll(t *testing.T) {
	t.Run("get all existing metrics", func(t *testing.T) {
		initialState := map[string]model.Metric{
			"gauge":   {ID: "g1", MType: model.Gauge, Value: floatp(10.5)},
			"counter": {ID: "c1", MType: model.Counter, Delta: intp(50)},
		}

		repo := inmemory.NewMemStorageMetricRepository()
		typedRepo := repo.(*inmemory.MemStorageMetricRepository)
		typedRepo.DB = initialState

		metricService := NewMetricService(typedRepo)

		allMetrics, err := metricService.GetAll()
		require.NoError(t, err)
		assert.Equal(t, 2, len(allMetrics.Metrics))

		foundNames := make(map[string]struct{})
		for _, m := range allMetrics.Metrics {
			foundNames[m.Name] = struct{}{}

			switch m.Name {
			case "gauge":
				assert.Equal(t, "10.5", m.Value)
			case "counter":
				assert.Equal(t, "50", m.Value)
			}
		}

		assert.Contains(t, foundNames, "gauge")
		assert.Contains(t, foundNames, "counter")
	})

	t.Run("get empty metrics map", func(t *testing.T) {
		initialState := map[string]model.Metric{}

		repo := inmemory.NewMemStorageMetricRepository()
		typedRepo := repo.(*inmemory.MemStorageMetricRepository)
		typedRepo.DB = initialState

		metricService := NewMetricService(typedRepo)

		allMetrics, err := metricService.GetAll()
		require.NoError(t, err)
		assert.Empty(t, allMetrics)
	})
}

func TestMetricServiceUpdate(t *testing.T) {
	type want struct {
		metric model.Metric
		exists bool
	}

	tests := []struct {
		name         string
		initialState map[string]model.Metric
		metricType   string
		metricName   string
		metricValue  string
		want         want
	}{
		{
			name:         "create Gauge",
			initialState: map[string]model.Metric{},
			metricType:   model.Gauge,
			metricName:   "gauge",
			metricValue:  "10.3",
			want: want{
				metric: model.Metric{MType: model.Gauge, Value: floatp(10.3)},
				exists: true,
			},
		},
		{
			name: "update Gauge",
			initialState: map[string]model.Metric{
				"existing_gauge": {ID: "uuid1", MType: model.Gauge, Value: floatp(9.1)},
			},
			metricType:  model.Gauge,
			metricName:  "existing_gauge",
			metricValue: "2.1",
			want: want{
				metric: model.Metric{ID: "uuid1", MType: model.Gauge, Value: floatp(2.1)},
				exists: true,
			},
		},
		{
			name:         "create Counter",
			initialState: map[string]model.Metric{},
			metricType:   model.Counter,
			metricName:   "new_counter",
			metricValue:  "10",
			want: want{
				metric: model.Metric{MType: model.Counter, Delta: intp(10)},
				exists: true,
			},
		},
		{
			name: "update Counter",
			initialState: map[string]model.Metric{
				"existing_counter": {ID: "uuid1", MType: model.Counter, Delta: intp(10)},
			},
			metricType:  model.Counter,
			metricName:  "existing_counter",
			metricValue: "10",
			want: want{
				metric: model.Metric{ID: "uuid1", MType: model.Counter, Delta: intp(20)},
				exists: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := inmemory.NewMemStorageMetricRepository()
			typedMockRepo, ok := mockRepo.(*inmemory.MemStorageMetricRepository)
			require.True(t, ok, "Failed to cast mockRepo to *MemStorageMetricRepository")

			typedMockRepo.DB = tt.initialState

			service := NewMetricService(mockRepo)
			err := service.Update(tt.metricType, tt.metricName, tt.metricValue)
			require.NoError(t, err)

			resultMetric, exists := mockRepo.GetByID(tt.metricName)

			require.Equal(t, tt.want.exists, exists)

			assert.Equal(t, tt.want.metric.MType, resultMetric.MType)
			assert.Equal(t, tt.want.metric.Value, resultMetric.Value)
			assert.Equal(t, tt.want.metric.Delta, resultMetric.Delta)
		})
	}
}
