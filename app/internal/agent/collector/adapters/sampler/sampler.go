package sampler

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
)

// MetricsSampler implements port.MetricsSampler.
type MetricsSampler struct{}

var _ port.MetricsSampler = (*MetricsSampler)(nil)

// NewMetricsSampler creates a new metrics sampler.
func NewMetricsSampler() *MetricsSampler {
	return &MetricsSampler{}
}

// ReadMemStats reads the current process memory statistics.
func (*MetricsSampler) ReadMemStats() runtime.MemStats {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms
}

// ReadVirtualMemory reads the total and free RAM.
func (*MetricsSampler) ReadVirtualMemory() (total, free uint64, err error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}
	return v.Total, v.Free, nil
}

// ReadCPUPercent reads the per-CPU utilization.
func (*MetricsSampler) ReadCPUPercent(interval time.Duration, perCPU bool) ([]float64, error) {
	return cpu.Percent(interval, perCPU)
}
