package port

import "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/domain/entity"

// MetricsRepository holds collected metrics state.
type MetricsRepository interface {
	UpdateRuntimeStats(randFloat func() float64)
	UpdateGopsutilStats(totalMem, freeMem float64, cpu []float64)
	Snapshot() entity.SystemState
}
