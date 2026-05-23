package inmemory

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricMemoryRepository_createAndGet(t *testing.T) {
	ctx := context.Background()
	repo := NewMetricMemoryRepository()
	v := 10.5
	m := entity.Metric{ID: "x", MType: entity.Gauge, Value: &v}

	require.NoError(t, repo.Create(ctx, m))

	got, err := repo.GetByID(ctx, "x")
	require.NoError(t, err)
	assert.Equal(t, v, *got.Value)

	_, err = repo.GetByID(ctx, "missing")
	assert.ErrorIs(t, err, application.ErrNotFound)
}

func TestMetricMemoryRepository_Ping(t *testing.T) {
	repo := NewMetricMemoryRepository()
	assert.NoError(t, repo.Ping(context.Background()))
}

func TestMetricMemoryRepository_batchAndQueries(t *testing.T) {
	ctx := context.Background()
	repo := NewMetricMemoryRepository()
	v1, v2 := 1.0, 2.0

	require.NoError(t, repo.CreateBatchWithParams(ctx, []entity.Metric{
		{ID: "a", MType: entity.Gauge, Value: &v1},
		{ID: "b", MType: entity.Gauge, Value: &v2},
	}))

	all, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)

	byIDs, err := repo.GetByIDs(ctx, []string{"a", "missing"})
	require.NoError(t, err)
	assert.Len(t, byIDs, 1)

	v3 := 3.0
	require.NoError(t, repo.Update(ctx, entity.Metric{ID: "a", MType: entity.Gauge, Value: &v3}))
	got, err := repo.GetByID(ctx, "a")
	require.NoError(t, err)
	assert.Equal(t, v3, *got.Value)
}

func TestMetricMemoryRepository_createDuplicate(t *testing.T) {
	ctx := context.Background()
	repo := NewMetricMemoryRepository()
	v := 1.0
	m := entity.Metric{ID: "x", MType: entity.Gauge, Value: &v}
	require.NoError(t, repo.Create(ctx, m))
	assert.ErrorIs(t, repo.Create(ctx, m), application.ErrAlreadyExists)
}
