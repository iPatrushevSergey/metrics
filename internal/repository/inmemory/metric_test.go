package inmemory

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func floatp(f float64) *float64 { return &f }
func intp(i int64) *int64       { return &i }

func TestMemStorageMetricRepository(t *testing.T) {
	ctx := context.Background()
	t.Run("create", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()

		mName := "testCreateGauge"
		mValue := 123.5

		createMetric := model.Metric{
			ID:    mName,
			MType: model.Gauge,
			Value: &mValue,
		}
		err := repo.Create(ctx, createMetric)
		require.NoError(t, err)

		repoMetric, err := repo.GetByID(ctx, mName)
		require.NoError(t, err)
		assert.Equal(t, createMetric, repoMetric)
	})

	t.Run("create duplicate returns error", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()
		mName := "testDuplicate"

		val1 := 10.0
		err := repo.Create(ctx, model.Metric{ID: mName, MType: model.Gauge, Value: &val1})
		require.NoError(t, err)

		val2 := 20.0
		err = repo.Create(ctx, model.Metric{ID: mName, MType: model.Gauge, Value: &val2})
		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrAlreadyExists)
	})

	t.Run("update existing metric (gauge)", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()
		mName := "testGauge"

		val1 := 10.0
		err := repo.Create(ctx, model.Metric{ID: mName, MType: model.Gauge, Value: &val1})
		require.NoError(t, err)

		val2 := 20.0
		updatedMetric := model.Metric{ID: mName, MType: model.Gauge, Value: &val2}
		err = repo.Update(ctx, mName, updatedMetric)
		require.NoError(t, err)

		repoMetric, err := repo.GetByID(ctx, mName)
		require.NoError(t, err)
		assert.Equal(t, val2, *repoMetric.Value)
	})

	t.Run("update existing metric (counter)", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()
		mName := "testCounter"

		delta1 := int64(10)
		err := repo.Create(ctx, model.Metric{ID: mName, MType: model.Counter, Delta: &delta1})
		require.NoError(t, err)

		delta2 := int64(100)
		updatedMetric := model.Metric{ID: mName, MType: model.Counter, Delta: &delta2}
		err = repo.Update(ctx, mName, updatedMetric)
		require.NoError(t, err)

		repoMetric, err := repo.GetByID(ctx, mName)
		require.NoError(t, err)
		assert.Equal(t, delta2, *repoMetric.Delta)
	})

	t.Run("update non-existent metric returns error", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()
		mName := "nonexistent"

		val := 10.0
		err := repo.Update(ctx, mName, model.Metric{ID: mName, MType: model.Gauge, Value: &val})
		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("get non-existent metric", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()

		_, err := repo.GetByID(ctx, "nonexistent")
		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("get all", func(t *testing.T) {
		initialState := map[string]model.Metric{
			"gauge":   {ID: "g1", MType: model.Gauge, Value: floatp(10.1)},
			"counter": {ID: "c1", MType: model.Counter, Delta: intp(5)},
		}

		repo := NewMemStorageMetricRepository()
		typedRepo := repo.(*MemStorageMetricRepository)

		typedRepo.DB = initialState

		allMetrics, err := typedRepo.GetAll(ctx)
		require.NoError(t, err)

		assert.Equal(t, 2, len(allMetrics))
		assert.Equal(t, initialState["gauge"], allMetrics["gauge"])
		assert.Equal(t, initialState["counter"], allMetrics["counter"])

		allMetrics["test"] = model.Metric{ID: "1", MType: model.Gauge, Value: floatp(11.1)}
		assert.Equal(t, 2, len(typedRepo.DB))
	})

	t.Run("concurrency safety check", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()

		var wg sync.WaitGroup
		numGoroutines := 10

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()

				gaugeName := fmt.Sprintf("gauge_%d", id)
				counterName := fmt.Sprintf("counter_%d", id)

				gaugeValue := 12.5
				counterValue := int64(15)

				_ = repo.Create(ctx, model.Metric{ID: gaugeName, MType: model.Gauge, Value: &gaugeValue})
				_ = repo.Create(ctx, model.Metric{ID: counterName, MType: model.Counter, Delta: &counterValue})

				_, _ = repo.GetByID(ctx, gaugeName)
				_, _ = repo.GetByID(ctx, counterName)
			}(i)
		}
		wg.Wait()
	})

	t.Run("GetByIDs", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()
		v := 1.5
		d := int64(10)
		_ = repo.Create(ctx, model.Metric{ID: "a", MType: model.Gauge, Value: &v})
		_ = repo.Create(ctx, model.Metric{ID: "b", MType: model.Counter, Delta: &d})
		byIDs, err := repo.GetByIDs(ctx, []string{"a", "b", "missing"})
		require.NoError(t, err)
		assert.Len(t, byIDs, 2)
		assert.Contains(t, byIDs, "a")
		assert.Contains(t, byIDs, "b")
	})

	t.Run("CreateBatch", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()
		v1, v2 := 1.0, 2.0
		batch := []model.Metric{
			{ID: "g1", MType: model.Gauge, Value: &v1},
			{ID: "g2", MType: model.Gauge, Value: &v2},
		}
		err := repo.CreateBatch(ctx, batch)
		require.NoError(t, err)
		all, _ := repo.GetAll(ctx)
		assert.Len(t, all, 2)
	})

	t.Run("UpdateBatch", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()
		v := 1.0
		_ = repo.Create(ctx, model.Metric{ID: "g1", MType: model.Gauge, Value: &v})
		v2 := 2.0
		err := repo.UpdateBatch(ctx, []model.Metric{{ID: "g1", MType: model.Gauge, Value: &v2}})
		require.NoError(t, err)
		m, _ := repo.GetByID(ctx, "g1")
		assert.Equal(t, 2.0, *m.Value)
	})
}

const benchmarkMetricsPerType = 100

// fillRepoForBench fills the repository with the initial data for the benchmark.
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

func BenchmarkMemStorage_GetByID(b *testing.B) {
	ctx := context.Background()
	repo := NewMemStorageMetricRepository()
	fillRepoForBench(ctx, repo, benchmarkMetricsPerType)
	targetID := "gauge_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, targetID)
	}
}

func BenchmarkMemStorage_GetAll(b *testing.B) {
	ctx := context.Background()
	repo := NewMemStorageMetricRepository()
	fillRepoForBench(ctx, repo, benchmarkMetricsPerType)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetAll(ctx)
	}
}

func BenchmarkMemStorage_Update(b *testing.B) {
	ctx := context.Background()
	repo := NewMemStorageMetricRepository()
	fillRepoForBench(ctx, repo, benchmarkMetricsPerType)
	v := 99.9
	m := model.Metric{ID: "gauge_0", MType: model.Gauge, Value: &v}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Update(ctx, "gauge_0", m)
	}
}

func BenchmarkMemStorage_CreateBatch(b *testing.B) {
	ctx := context.Background()
	metrics := make([]model.Metric, 0, 50)
	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("batch_g_%d", i)
		v := float64(i)
		metrics = append(metrics, model.Metric{ID: name, MType: model.Gauge, Value: &v})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo := NewMemStorageMetricRepository()
		_ = repo.CreateBatch(ctx, metrics)
	}
}

func BenchmarkMemStorage_UpdateBatch(b *testing.B) {
	ctx := context.Background()
	repo := NewMemStorageMetricRepository()
	fillRepoForBench(ctx, repo, 25)
	metrics := make([]model.Metric, 0, 25)
	for i := 0; i < 25; i++ {
		name := fmt.Sprintf("gauge_%d", i)
		v := float64(i + 100)
		metrics = append(metrics, model.Metric{ID: name, MType: model.Gauge, Value: &v})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.UpdateBatch(ctx, metrics)
	}
}
