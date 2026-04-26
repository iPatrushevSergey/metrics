package sampler

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
)

// StatsSampler implements port.StatsSampler.
type StatsSampler struct{}

var _ port.StatsSampler = (*StatsSampler)(nil)

// NewStatsSampler creates a new stats sampler.
func NewStatsSampler() *StatsSampler {
	return &StatsSampler{}
}

// ReadMemStats reads the current process memory stats.
func (*StatsSampler) ReadMemStats() runtime.MemStats {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms
}

// ReadVirtualMemory reads the total and free RAM.
func (*StatsSampler) ReadVirtualMemory() (total, free uint64, err error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}
	return v.Total, v.Free, nil
}

// ReadCPUPercent reads the per-CPU utilization.
func (*StatsSampler) ReadCPUPercent(interval time.Duration, perCPU bool) ([]float64, error) {
	return cpu.Percent(interval, perCPU)
}
