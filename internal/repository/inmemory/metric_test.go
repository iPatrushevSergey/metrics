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
}
