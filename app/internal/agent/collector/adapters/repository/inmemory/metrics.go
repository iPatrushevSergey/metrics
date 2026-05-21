package inmemory

import (
	"runtime"
	"sync"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/domain/entity"
)

// MetricsRepository repository for collected metrics state.
type MetricsRepository struct {
	mu sync.RWMutex
	st entity.SystemMetrics
}

var _ port.MetricsRepository = (*MetricsRepository)(nil)

// NewMetricsRepository returns a new metrics repository.
func NewMetricsRepository() *MetricsRepository {
	return &MetricsRepository{}
}

// UpdateRuntimeMetrics updates runtime memstats and poll stats.
func (s *MetricsRepository) UpdateRuntimeMetrics(ms runtime.MemStats, randValue float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.st.Runtime.MemStats = ms
	s.st.Poll.PollCount++
	s.st.Poll.RandomValue = randValue
}

// UpdateGopsutilMetrics updates host memory and CPU samples.
func (s *MetricsRepository) UpdateGopsutilMetrics(totalMem, freeMem float64, cpu []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.st.Gopsutil.TotalMemory = totalMem
	s.st.Gopsutil.FreeMemory = freeMem
	s.st.Gopsutil.CPUutilization = cpu
}

// GetSystemState returns the current system state
func (s *MetricsRepository) GetSystemMetrics() entity.SystemMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return entity.SystemMetrics{
		Runtime: s.st.Runtime,
		Poll:    s.st.Poll,
		Gopsutil: entity.GopsutilMetrics{
			TotalMemory:    s.st.Gopsutil.TotalMemory,
			FreeMemory:     s.st.Gopsutil.FreeMemory,
			CPUutilization: append([]float64(nil), s.st.Gopsutil.CPUutilization...),
		},
	}
}
