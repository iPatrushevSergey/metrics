package system

import (
	"runtime"
	"sync"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/domain/entity"
)

// MetricsRepository repository for collected metrics state.
type MetricsRepository struct {
	mu sync.RWMutex
	st entity.SystemState
}

var _ port.MetricsRepository = (*MetricsRepository)(nil)

// NewMetricsRepository returns a new metrics repository.
func NewMetricsRepository() *MetricsRepository {
	return &MetricsRepository{}
}

// UpdateRuntimeStats updates runtime memstats and poll stats.
func (s *MetricsRepository) UpdateRuntimeStats(randFloat func() float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	runtime.ReadMemStats(&s.st.MemStats)
	s.st.Poll.PollCount++
	s.st.Poll.RandomValue = randFloat()
}

// UpdateGopsutilStats updates gopsutil stats.
func (s *MetricsRepository) UpdateGopsutilStats(totalMem, freeMem float64, cpu []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.st.Gopsutil.TotalMemory = totalMem
	s.st.Gopsutil.FreeMemory = freeMem
	s.st.Gopsutil.CPUutilization = cpu
}

// Snapshot returns metrics of the current system status.
func (s *MetricsRepository) Snapshot() entity.SystemState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return entity.SystemState{
		MemStats: s.st.MemStats,
		Poll:     s.st.Poll,
		Gopsutil: entity.GopsutilStats{
			TotalMemory:    s.st.Gopsutil.TotalMemory,
			FreeMemory:     s.st.Gopsutil.FreeMemory,
			CPUutilization: append([]float64(nil), s.st.Gopsutil.CPUutilization...),
		},
	}
}
