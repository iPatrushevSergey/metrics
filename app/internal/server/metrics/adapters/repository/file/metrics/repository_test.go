package metrics

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricFileRepository_saveAndLoad(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metrics.json")
	repo := NewMetricFileRepository(path)

	v := 10.0
	metrics := map[string]entity.Metric{
		"x": {ID: "x", MType: entity.Gauge, Value: &v},
	}
	require.NoError(t, repo.SaveAll(ctx, metrics))

	loaded, err := repo.LoadAll(ctx)
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	assert.Equal(t, "x", loaded[0].ID)
	require.NotNil(t, loaded[0].Value)
	assert.Equal(t, v, *loaded[0].Value)
}

func TestMetricFileRepository_loadEmpty(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "empty.json")
	repo := NewMetricFileRepository(path)

	loaded, err := repo.LoadAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, loaded)
}
