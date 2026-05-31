package sampler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsSampler_ReadMemStats(t *testing.T) {
	s := NewMetricsSampler()
	ms := s.ReadMemStats()
	assert.NotZero(t, ms.Sys)
}

func TestMetricsSampler_ReadVirtualMemory(t *testing.T) {
	s := NewMetricsSampler()
	total, free, err := s.ReadVirtualMemory()
	require.NoError(t, err)
	assert.NotZero(t, total)
	assert.LessOrEqual(t, free, total)
}

func TestMetricsSampler_ReadCPUPercent(t *testing.T) {
	s := NewMetricsSampler()
	percents, err := s.ReadCPUPercent(10*time.Millisecond, true)
	require.NoError(t, err)
	assert.NotEmpty(t, percents)
}
