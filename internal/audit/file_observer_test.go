package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFileObserver_writesLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	obs, err := NewFileObserver(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = obs.Close() })

	e := Event{TS: 42, Metrics: []string{"m1"}, IPAddress: "127.0.0.1"}
	require.NoError(t, obs.Publish(context.Background(), e))

	require.NoError(t, obs.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), `"ts":42`)
}

func TestFileObserver_Publish_afterClose(t *testing.T) {
	dir := t.TempDir()
	obs, err := NewFileObserver(filepath.Join(dir, "a.log"))
	require.NoError(t, err)
	require.NoError(t, obs.Close())

	err = obs.Publish(context.Background(), Event{})
	require.ErrorIs(t, err, errFileObserverClosed)
}

func TestFileObserver_Publish_cancelledContext(t *testing.T) {
	dir := t.TempDir()
	obs, err := NewFileObserver(filepath.Join(dir, "b.log"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = obs.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = obs.Publish(ctx, Event{TS: 1})
	require.ErrorIs(t, err, context.Canceled)
}

func TestFileObserver_Close_idempotent(t *testing.T) {
	dir := t.TempDir()
	obs, err := NewFileObserver(filepath.Join(dir, "c.log"))
	require.NoError(t, err)
	require.NoError(t, obs.Close())
	require.NoError(t, obs.Close())
}
