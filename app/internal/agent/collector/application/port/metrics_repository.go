package port

import (
	"runtime"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/domain/entity"
)

// MetricsRepository holds collected metrics state.
type MetricsRepository interface {
	UpdateRuntimeMetrics(ms runtime.MemStats, randValue float64)
	UpdateGopsutilMetrics(totalMem, freeMem float64, cpu []float64)
	GetSystemMetrics() entity.SystemMetrics
}
