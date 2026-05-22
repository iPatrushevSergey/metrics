package service

import (
	"runtime"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestCountersFromState(t *testing.T) {
	counters := CountersFromState(&entity.AgentPollMetrics{PollCount: 7})

	assert.Equal(t, int64(7), counters["PollCount"])
}

func TestGaugesFromState(t *testing.T) {
	ms := runtime.MemStats{Alloc: 100, Sys: 200}
	poll := &entity.AgentPollMetrics{RandomValue: 1.5}
	gs := &entity.GopsutilMetrics{
		TotalMemory:    1000,
		FreeMemory:     400,
		CPUutilization: []float64{11, 22},
	}

	gauges := GaugesFromState(&ms, poll, gs)

	assert.Equal(t, float64(100), gauges["Alloc"])
	assert.Equal(t, 1.5, gauges["RandomValue"])
	assert.Equal(t, 1000.0, gauges["TotalMemory"])
	assert.Equal(t, 11.0, gauges["CPUutilization1"])
	assert.Equal(t, 22.0, gauges["CPUutilization2"])
}
