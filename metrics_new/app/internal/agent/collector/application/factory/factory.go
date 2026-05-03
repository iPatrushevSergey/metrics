package factory

import (
	"time"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/agent/collector/application/usecase"
)

// Params contains dependencies required to build collector use cases.
type Params struct {
	MetricsRepo    port.MetricsRepository
	MetricsSampler port.MetricsSampler
	MetricsClient  port.MetricsClient
	Log            port.Logger
	RandFloat      func() float64
	ReportInterval time.Duration
}

// UseCases holds collector module use cases exposed to composition root.
type UseCases struct {
	PollRuntime  *usecase.PollRuntimeTick
	PollGopsutil *usecase.PollGopsutilTick
	ReportBatch  *usecase.ReportBatchTick
}

// NewUseCases builds collector module use cases.
func NewUseCases(p Params) *UseCases {
	return &UseCases{
		PollRuntime:  usecase.NewPollRuntimeTick(p.MetricsSampler, p.MetricsRepo, p.RandFloat),
		PollGopsutil: usecase.NewPollGopsutilTick(p.MetricsSampler, p.MetricsRepo, p.Log),
		ReportBatch:  usecase.NewReportBatchTick(p.MetricsRepo, p.MetricsClient, p.Log, p.ReportInterval),
	}
}
