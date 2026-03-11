package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func floatp(f float64) *float64 { return &f }
func intp(i int64) *int64       { return &i }

func TestMetricServiceGet(t *testing.T) {
	ctx := context.Background()

	initialState := map[string]model.Metric{
		"gauge":   {ID: "g1", MType: model.Gauge, Value: floatp(10.5)},
		"counter": {ID: "c1", MType: model.Counter, Delta: intp(50)},
	}

	repo := inmemory.NewMemStorageMetricRepository()
	typedRepo := repo.(*inmemory.MemStorageMetricRepository)
	typedRepo.DB = initialState

	metricService := NewMetricService(typedRepo)

	t.Run("success get metric", func(t *testing.T) {
		metric, err := metricService.GetValue(ctx, "gauge", "gauge")
		require.NoError(t, err)
		assert.Equal(t, "10.5", metric)
	})

	t.Run("error metric not found", func(t *testing.T) {
		_, err := metricService.GetValue(ctx, "counter", "nonexistent")
		require.Error(t, err)
		assert.EqualError(t, err, "metric not found")
	})
}

func TestMetricServiceGetAll(t *testing.T) {
	ctx := context.Background()

	t.Run("get all existing metrics", func(t *testing.T) {
		initialState := map[string]model.Metric{
			"gauge":   {ID: "g1", MType: model.Gauge, Value: floatp(10.5)},
			"counter": {ID: "c1", MType: model.Counter, Delta: intp(50)},
		}

		repo := inmemory.NewMemStorageMetricRepository()
		typedRepo := repo.(*inmemory.MemStorageMetricRepository)
		typedRepo.DB = initialState

		metricService := NewMetricService(typedRepo)

		allMetrics, err := metricService.GetAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, len(allMetrics))

		assert.Contains(t, allMetrics, "gauge")
		assert.Contains(t, allMetrics, "counter")
		assert.Equal(t, floatp(10.5), allMetrics["gauge"].Value)
		assert.Equal(t, intp(50), allMetrics["counter"].Delta)
	})

	t.Run("get empty metrics map", func(t *testing.T) {
		initialState := map[string]model.Metric{}

		repo := inmemory.NewMemStorageMetricRepository()
		typedRepo := repo.(*inmemory.MemStorageMetricRepository)
		typedRepo.DB = initialState

		metricService := NewMetricService(typedRepo)

		allMetrics, err := metricService.GetAll(ctx)
		require.NoError(t, err)
		assert.Empty(t, allMetrics)
	})
}

func TestMetricServiceUpdate(t *testing.T) {
	ctx := context.Background()

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
			metricType:   string(model.Gauge),
			metricName:   "gauge",
			metricValue:  "10.3",
			want: want{
				metric: model.Metric{ID: "gauge", MType: model.Gauge, Value: floatp(10.3)},
				exists: true,
			},
		},
		{
			name: "update Gauge",
			initialState: map[string]model.Metric{
				"existing_gauge": {ID: "existing_gauge", MType: model.Gauge, Value: floatp(9.1)},
			},
			metricType:  string(model.Gauge),
			metricName:  "existing_gauge",
			metricValue: "2.1",
			want: want{
				metric: model.Metric{ID: "existing_gauge", MType: model.Gauge, Value: floatp(2.1)},
				exists: true,
			},
		},
		{
			name:         "create Counter",
			initialState: map[string]model.Metric{},
			metricType:   string(model.Counter),
			metricName:   "new_counter",
			metricValue:  "10",
			want: want{
				metric: model.Metric{ID: "new_counter", MType: model.Counter, Delta: intp(10)},
				exists: true,
			},
		},
		{
			name: "update Counter",
			initialState: map[string]model.Metric{
				"existing_counter": {ID: "existing_counter", MType: model.Counter, Delta: intp(10)},
			},
			metricType:  string(model.Counter),
			metricName:  "existing_counter",
			metricValue: "10",
			want: want{
				metric: model.Metric{ID: "existing_counter", MType: model.Counter, Delta: intp(20)},
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
			err := service.Update(ctx, tt.metricType, tt.metricName, tt.metricValue)
			require.NoError(t, err)

			resultMetric, err := mockRepo.GetByID(ctx, tt.metricName)
			require.NoError(t, err)

			assert.Equal(t, tt.want.metric.MType, resultMetric.MType)
			assert.Equal(t, tt.want.metric.Value, resultMetric.Value)
			assert.Equal(t, tt.want.metric.Delta, resultMetric.Delta)
		})
	}
}

func TestFormatMetric(t *testing.T) {
	repo := inmemory.NewMemStorageMetricRepository()
	svc := NewMetricService(repo)
	v := 1.5
	d := int64(10)

	t.Run("gauge", func(t *testing.T) {
		s, err := svc.FormatMetric(model.Metric{MType: model.Gauge, Value: &v})
		require.NoError(t, err)
		assert.Equal(t, "1.5", s)
	})
	t.Run("counter", func(t *testing.T) {
		s, err := svc.FormatMetric(model.Metric{MType: model.Counter, Delta: &d})
		require.NoError(t, err)
		assert.Equal(t, "10", s)
	})
	t.Run("gauge nil value", func(t *testing.T) {
		_, err := svc.FormatMetric(model.Metric{MType: model.Gauge, Value: nil})
		require.Error(t, err)
	})
	t.Run("counter nil delta", func(t *testing.T) {
		_, err := svc.FormatMetric(model.Metric{MType: model.Counter, Delta: nil})
		require.Error(t, err)
	})
	t.Run("unknown type", func(t *testing.T) {
		_, err := svc.FormatMetric(model.Metric{MType: "unknown"})
		require.Error(t, err)
	})
}

func TestPingDB(t *testing.T) {
	repo := inmemory.NewMemStorageMetricRepository()
	svc := NewMetricService(repo)
	err := svc.PingDB(context.Background())
	require.NoError(t, err)
}

func TestMetricServiceUpdateJSON(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMemStorageMetricRepository()
	svc := NewMetricService(repo)

	t.Run("create gauge", func(t *testing.T) {
		v := 1.5
		err := svc.UpdateJSON(ctx, model.Metric{ID: "g1", MType: model.Gauge, Value: &v})
		require.NoError(t, err)
		val, err := svc.GetValue(ctx, string(model.Gauge), "g1")
		require.NoError(t, err)
		assert.Equal(t, "1.5", val)
	})
	t.Run("update gauge", func(t *testing.T) {
		v := 2.5
		err := svc.UpdateJSON(ctx, model.Metric{ID: "g1", MType: model.Gauge, Value: &v})
		require.NoError(t, err)
		val, err := svc.GetValue(ctx, string(model.Gauge), "g1")
		require.NoError(t, err)
		assert.Equal(t, "2.5", val)
	})
	t.Run("create and increment counter", func(t *testing.T) {
		d := int64(1)
		err := svc.UpdateJSON(ctx, model.Metric{ID: "c1", MType: model.Counter, Delta: &d})
		require.NoError(t, err)
		err = svc.UpdateJSON(ctx, model.Metric{ID: "c1", MType: model.Counter, Delta: &d})
		require.NoError(t, err)
		val, err := svc.GetValue(ctx, string(model.Counter), "c1")
		require.NoError(t, err)
		assert.Equal(t, "2", val)
	})
}

func TestMetricServiceUpdatesJSON(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMemStorageMetricRepository()
	svc := NewMetricService(repo)

	v1, v2 := 1.0, 2.0
	d1 := int64(10)
	metrics := []model.Metric{
		{ID: "g1", MType: model.Gauge, Value: &v1},
		{ID: "g2", MType: model.Gauge, Value: &v2},
		{ID: "c1", MType: model.Counter, Delta: &d1},
	}
	err := svc.UpdatesJSON(ctx, metrics)
	require.NoError(t, err)

	all, err := svc.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 3)
}

func TestMetricServiceUpdatesJSON_mergeDuplicates(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewMemStorageMetricRepository()
	svc := NewMetricService(repo)

	d1, d2 := int64(5), int64(3)
	metrics := []model.Metric{
		{ID: "c1", MType: model.Counter, Delta: &d1},
		{ID: "c1", MType: model.Counter, Delta: &d2},
	}
	err := svc.UpdatesJSON(ctx, metrics)
	require.NoError(t, err)
	val, err := svc.GetValue(ctx, string(model.Counter), "c1")
	require.NoError(t, err)
	assert.Equal(t, "8", val)
}

const serviceBenchMetricsPerType = 50

func fillRepoForBench(ctx context.Context, repo repository.MetricRepository, n int) {
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("gauge_%d", i)
		v := float64(i)
		_ = repo.Create(ctx, model.Metric{ID: name, MType: model.Gauge, Value: &v})
	}
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("counter_%d", i)
		d := int64(i)
		_ = repo.Create(ctx, model.Metric{ID: name, MType: model.Counter, Delta: &d})
	}
}

func BenchmarkMetricsService_GetValue(b *testing.B) {
	ctx := context.Background()
	repo := inmemory.NewMemStorageMetricRepository()
	fillRepoForBench(ctx, repo, serviceBenchMetricsPerType)
	svc := NewMetricService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetValue(ctx, string(model.Gauge), "gauge_0")
	}
}

func BenchmarkMetricsService_GetAll(b *testing.B) {
	ctx := context.Background()
	repo := inmemory.NewMemStorageMetricRepository()
	fillRepoForBench(ctx, repo, serviceBenchMetricsPerType)
	svc := NewMetricService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetAll(ctx)
	}
}

func BenchmarkMetricsService_Update(b *testing.B) {
	ctx := context.Background()
	repo := inmemory.NewMemStorageMetricRepository()
	fillRepoForBench(ctx, repo, serviceBenchMetricsPerType)
	svc := NewMetricService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.Update(ctx, string(model.Gauge), "gauge_0", "123.45")
	}
}

func BenchmarkMetricsService_UpdateJSON(b *testing.B) {
	ctx := context.Background()
	repo := inmemory.NewMemStorageMetricRepository()
	fillRepoForBench(ctx, repo, serviceBenchMetricsPerType)
	svc := NewMetricService(repo)
	v := 456.78
	m := model.Metric{ID: "gauge_0", MType: model.Gauge, Value: &v}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.UpdateJSON(ctx, m)
	}
}

func BenchmarkMetricsService_UpdatesJSON(b *testing.B) {
	ctx := context.Background()
	repo := inmemory.NewMemStorageMetricRepository()
	fillRepoForBench(ctx, repo, 20)
	svc := NewMetricService(repo)
	metrics := make([]model.Metric, 0, 30)
	for i := 0; i < 30; i++ {
		name := fmt.Sprintf("batch_%d", i)
		val := float64(i)
		metrics = append(metrics, model.Metric{ID: name, MType: model.Gauge, Value: &val})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.UpdatesJSON(ctx, metrics)
	}
}
