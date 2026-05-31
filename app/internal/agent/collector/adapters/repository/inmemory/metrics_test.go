package inmemory

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsRepository(t *testing.T) {
	repo := NewMetricsRepository()

	repo.UpdateRuntimeMetrics(runtime.MemStats{Alloc: 10}, 3.3)
	repo.UpdateGopsutilMetrics(100, 40, []float64{1, 2})

	snap := repo.GetSystemMetrics()

	assert.Equal(t, uint64(10), snap.Runtime.MemStats.Alloc)
	assert.Equal(t, int64(1), snap.Poll.PollCount)
	assert.Equal(t, 3.3, snap.Poll.RandomValue)
	assert.Equal(t, 100.0, snap.Gopsutil.TotalMemory)
	assert.Equal(t, []float64{1, 2}, snap.Gopsutil.CPUutilization)
}
