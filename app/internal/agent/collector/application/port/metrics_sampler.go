package port

import (
	"runtime"
	"time"
)

// MetricsSampler reads runtime and OS inputs used to fill the metrics snapshot.
type MetricsSampler interface {
	ReadMemStats() runtime.MemStats
	ReadVirtualMemory() (total, free uint64, err error)
	ReadCPUPercent(interval time.Duration, perCPU bool) ([]float64, error)
}
