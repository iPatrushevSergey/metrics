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
		metric, err := metricService.Get("gauge")
		require.NoError(t, err)
		assert.Equal(t, floatp(10.5), metric.Value)
	})

	t.Run("error metric not found", func(t *testing.T) {
		_, err := metricService.Get("nonexistent")
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

		allMetrics := metricService.GetAll()
		assert.Equal(t, 2, len(allMetrics))
		assert.Contains(t, allMetrics, "gauge")
		assert.Contains(t, allMetrics, "counter")
	})

	t.Run("get empty metrics map", func(t *testing.T) {
		initialState := map[string]model.Metric{}

		repo := inmemory.NewMemStorageMetricRepository()
		typedRepo := repo.(*inmemory.MemStorageMetricRepository)
		typedRepo.DB = initialState

		metricService := NewMetricService(typedRepo)

		allMetrics := metricService.GetAll()
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
		metricValue  any
		want         want
	}{
		{
			name:         "create Gauge",
			initialState: map[string]model.Metric{},
			metricType:   model.Gauge,
			metricName:   "newGauge",
			metricValue:  float64(10.3),
			want: want{
				metric: model.Metric{MType: model.Gauge, Value: floatp(10.3)},
				exists: true,
			},
		},
		{
			name: "update Gauge",
			initialState: map[string]model.Metric{
				"existingGauge": {ID: "uuid1", MType: model.Gauge, Value: floatp(9.1)},
			},
			metricType:  model.Gauge,
			metricName:  "existingGauge",
			metricValue: float64(2.1),
			want: want{
				metric: model.Metric{ID: "uuid1", MType: model.Gauge, Value: floatp(2.1)},
				exists: true,
			},
		},
		{
			name:         "create Counter",
			initialState: map[string]model.Metric{},
			metricType:   model.Counter,
			metricName:   "newCounter",
			metricValue:  int64(10),
			want: want{
				metric: model.Metric{MType: model.Counter, Delta: intp(10)},
				exists: true,
			},
		},
		{
			name: "update Counter",
			initialState: map[string]model.Metric{
				"existingCounter": {ID: "uuid1", MType: model.Counter, Delta: intp(10)},
			},
			metricType:  model.Counter,
			metricName:  "existingCounter",
			metricValue: int64(10),
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

			resultMetric, exists := mockRepo.GetByName(tt.metricName)

			require.Equal(t, tt.want.exists, exists)

			assert.Equal(t, tt.want.metric.MType, resultMetric.MType)
			assert.Equal(t, tt.want.metric.Value, resultMetric.Value)
			assert.Equal(t, tt.want.metric.Delta, resultMetric.Delta)
		})
	}
}
