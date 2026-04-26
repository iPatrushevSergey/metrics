package port

import (
	"runtime"
	"time"
)

// StatsSampler reads runtime and OS metrics.
type StatsSampler interface {
	ReadMemStats() runtime.MemStats
	ReadVirtualMemory() (total, free uint64, err error)
	ReadCPUPercent(interval time.Duration, perCPU bool) ([]float64, error)
}
