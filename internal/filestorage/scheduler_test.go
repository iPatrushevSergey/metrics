package filestorage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/stretchr/testify/require"
)

func TestNewPeriodicSaver_Start_Stop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "m.json")
	fs := NewFileStorage(path)
	repo := inmemory.NewMemStorageMetricRepository()
	ps := NewPeriodicSaver(repo, fs, 10*time.Hour)
	ps.Start()
	ps.Stop()
}

func TestSaveOnShutdown(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "m.json")
	fs := NewFileStorage(path)
	repo := inmemory.NewMemStorageMetricRepository()
	v := 1.5
	_ = repo.Create(ctx, model.Metric{ID: "g", MType: model.Gauge, Value: &v})

	err := SaveOnShutdown(ctx, repo, fs)
	require.NoError(t, err)

	loaded, err := fs.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 1)
}
