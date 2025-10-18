package inmemory

import (
	"fmt"
	"sync"
	"testing"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorageMetricRepository(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()

		mName := "testCreateGauge"
		mValue := 123.5

		createMetric := model.Metric{
			ID:    "testID",
			MType: model.Gauge,
			Value: &mValue,
		}
		repo.Create(mName, createMetric)

		repoMetric, exists := repo.GetByName(mName)
		require.True(t, exists)
		assert.Equal(t, createMetric, repoMetric)
	})

	t.Run("update", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()

		mName := "testUpdateCounter"
		mDelta1 := int64(10)
		mDelta2 := int64(100)

		createMetric := model.Metric{
			ID:    "testID",
			MType: model.Counter,
			Delta: &mDelta1,
		}
		repo.Create(mName, createMetric)

		updateMetric := model.Metric{
			ID:    "id",
			MType: model.Counter,
			Delta: &mDelta2,
		}
		repo.Update(mName, updateMetric)

		repoMetric, exists := repo.GetByName(mName)
		require.True(t, exists)
		assert.Equal(t, updateMetric, repoMetric)
	})

	t.Run("get non-existent metric", func(t *testing.T) {
		repo := NewMemStorageMetricRepository()

		_, exists := repo.GetByName("nonexistent")
		assert.False(t, exists)
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

				repo.Create(gaugeName, model.Metric{Value: &gaugeValue})
				repo.Create(counterName, model.Metric{Delta: &counterValue})

				_, _ = repo.GetByName(gaugeName)
				_, _ = repo.GetByName(counterName)
			}()
		}
		wg.Wait()
	})
}
