package filestorage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileStorage(t *testing.T) {
	fs := NewFileStorage("/tmp/metrics.json")
	assert.Equal(t, "/tmp/metrics.json", fs.FilePath)
}

func TestFileStorage_Save_Load(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")
	fs := NewFileStorage(path)

	v := 1.5
	d := int64(10)
	metrics := map[string]model.Metric{
		"g": {ID: "g", MType: model.Gauge, Value: &v},
		"c": {ID: "c", MType: model.Counter, Delta: &d},
	}

	err := fs.Save(metrics)
	require.NoError(t, err)

	loaded, err := fs.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 2)
}

func TestFileStorage_Load_emptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	require.NoError(t, os.WriteFile(path, []byte("[]"), 0666))
	fs := NewFileStorage(path)

	loaded, err := fs.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded)
}
