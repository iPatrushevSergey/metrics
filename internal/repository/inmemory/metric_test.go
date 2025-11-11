package inmemory

import (
	"fmt"
	"sync"
	"testing"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func floatp(f float64) *float64 { return &f }
func intp(i int64) *int64       { return &i }

func TestMemStorageMetricRepository(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()

		mName := "testCreateGauge"
		mValue := 123.5

		createMetric := model.Metric{
			ID:    mName,
			MType: model.Gauge,
			Value: &mValue,
		}
		repo.Create(createMetric)

		repoMetric, exists := repo.GetByID(mName)
		require.True(t, exists)
		assert.Equal(t, createMetric, repoMetric)
	})

	t.Run("update", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()

		mName := "testUpdateCounter"
		mDelta1 := int64(10)
		mDelta2 := int64(100)

		createMetric := model.Metric{
			ID:    mName,
			MType: model.Counter,
			Delta: &mDelta1,
		}
		repo.Create(createMetric)

		updateMetric := model.Metric{
			ID:    mName,
			MType: model.Counter,
			Delta: &mDelta2,
		}
		repo.Update(mName, updateMetric)

		repoMetric, exists := repo.GetByID(mName)
		require.True(t, exists)
		assert.Equal(t, updateMetric, repoMetric)
	})

	t.Run("get non-existent metric", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()

		_, exists := repo.GetByID("nonexistent")
		assert.False(t, exists)
	})

	t.Run("get all", func(t *testing.T) {
		initialState := map[string]model.Metric{
			"gauge":   {ID: "g1", MType: model.Gauge, Value: floatp(10.1)},
			"counter": {ID: "c1", MType: model.Counter, Delta: intp(5)},
		}

		repo := NewMemStorageMetricRepository()
		typedRepo := repo.(*MemStorageMetricRepository)

		typedRepo.DB = initialState

		allMetrics := typedRepo.GetAll()
		assert.Equal(t, 2, len(initialState))
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
			go func() {
				defer wg.Done()

				gaugeName := fmt.Sprintf("gauge_%d", i)
				counterName := fmt.Sprintf("counter_%d", i)

				gaugeValue := 12.5
				counterValue := int64(15)

				repo.Create(model.Metric{ID: gaugeName, Value: &gaugeValue})
				repo.Create(model.Metric{ID: counterName, Delta: &counterValue})

				_, _ = repo.GetByID(gaugeName)
				_, _ = repo.GetByID(counterName)
			}()
		}
		wg.Wait()
	})
}
